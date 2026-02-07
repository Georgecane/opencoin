package p2p

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"

	"github.com/georgecane/opencoin/pkg/crypto"
)

const kyberProtocol = protocol.ID("/opencoin/kyber/1.0")

// Config defines P2P configuration.
type Config struct {
	ListenAddrs    []string
	BootstrapPeers []string
}

// P2P provides libp2p host + pubsub + Kyber key exchange.
type P2P struct {
	ctx       context.Context
	Host      host.Host
	DHT       *dht.IpfsDHT
	PubSub    *pubsub.PubSub
	sessionMu sync.RWMutex
	sessions  map[peer.ID][]byte
}

// New creates a new P2P host.
func New(ctx context.Context, cfg Config) (*P2P, error) {
	opts := []libp2p.Option{}
	for _, addr := range cfg.ListenAddrs {
		opts = append(opts, libp2p.ListenAddrStrings(addr))
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("libp2p new: %w", err)
	}
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("pubsub: %w", err)
	}
	dhtNode, err := dht.New(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("dht: %w", err)
	}
	p := &P2P{
		ctx:      ctx,
		Host:     h,
		DHT:      dhtNode,
		PubSub:   ps,
		sessions: make(map[peer.ID][]byte),
	}
	h.SetStreamHandler(kyberProtocol, p.handleKyberStream)

	// Connect to bootstrap peers.
	for _, addr := range cfg.BootstrapPeers {
		if err := p.connectPeer(addr); err != nil {
			return nil, err
		}
	}
	// Trigger Kyber handshake on new connections.
	h.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(_ network.Network, conn network.Conn) {
			go p.initiateKyber(conn.RemotePeer())
		},
	})
	return p, nil
}

func (p *P2P) connectPeer(addr string) error {
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("invalid peer addr: %w", err)
	}
	info, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return fmt.Errorf("addr info: %w", err)
	}
	return p.Host.Connect(p.ctx, *info)
}

// Topic returns a pubsub topic.
func (p *P2P) Topic(name string) (*pubsub.Topic, error) {
	return p.PubSub.Join(name)
}

// SessionKey returns the Kyber-derived session key for a peer.
func (p *P2P) SessionKey(id peer.ID) []byte {
	p.sessionMu.RLock()
	defer p.sessionMu.RUnlock()
	return append([]byte(nil), p.sessions[id]...)
}

// EncryptForPeer encrypts payload using the peer session key.
func (p *P2P) EncryptForPeer(id peer.ID, plaintext []byte) ([]byte, error) {
	key := p.SessionKey(id)
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("missing session key")
	}
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ct := aead.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ct...), nil
}

// DecryptFromPeer decrypts payload using the peer session key.
func (p *P2P) DecryptFromPeer(id peer.ID, ciphertext []byte) ([]byte, error) {
	key := p.SessionKey(id)
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("missing session key")
	}
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aead.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:aead.NonceSize()]
	ct := ciphertext[aead.NonceSize():]
	return aead.Open(nil, nonce, ct, nil)
}

func (p *P2P) initiateKyber(peerID peer.ID) {
	stream, err := p.Host.NewStream(p.ctx, peerID, kyberProtocol)
	if err != nil {
		return
	}
	defer stream.Close()

	kp, err := crypto.GenerateKeyPairKyber768()
	if err != nil {
		return
	}
	if err := writeBytes(stream, kp.PublicKey); err != nil {
		return
	}
	ct, err := readBytes(stream)
	if err != nil {
		return
	}
	ss, err := crypto.DecapsulateKyber768(kp.PrivateKey, ct)
	if err != nil {
		return
	}
	p.storeSession(peerID, ss)
}

func (p *P2P) handleKyberStream(stream network.Stream) {
	defer stream.Close()
	pk, err := readBytes(stream)
	if err != nil {
		return
	}
	ct, ss, err := crypto.EncapsulateKyber768(pk)
	if err != nil {
		return
	}
	if err := writeBytes(stream, ct); err != nil {
		return
	}
	p.storeSession(stream.Conn().RemotePeer(), ss)
}

func (p *P2P) storeSession(peerID peer.ID, secret []byte) {
	// Derive a deterministic 32-byte session key.
	h := hkdf.New(sha256.New, secret, nil, []byte("opencoin-kyber"))
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(h, key); err != nil {
		return
	}
	p.sessionMu.Lock()
	p.sessions[peerID] = key
	p.sessionMu.Unlock()
}

func writeBytes(w io.Writer, b []byte) error {
	if len(b) > 1<<20 {
		return fmt.Errorf("message too large")
	}
	lenBuf := []byte{byte(len(b) >> 8), byte(len(b))}
	if _, err := w.Write(lenBuf); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

func readBytes(r io.Reader) ([]byte, error) {
	lenBuf := make([]byte, 2)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}
	size := int(lenBuf[0])<<8 | int(lenBuf[1])
	if size <= 0 || size > 1<<20 {
		return nil, fmt.Errorf("invalid size")
	}
	out := make([]byte, size)
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, err
	}
	return out, nil
}

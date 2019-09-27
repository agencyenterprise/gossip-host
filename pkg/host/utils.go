package host

import (
	"context"
	"strings"

	"github.com/agencyenterprise/gossip-host/pkg/logger"

	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	mplex "github.com/libp2p/go-libp2p-mplex"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	quic "github.com/libp2p/go-libp2p-quic-transport"
	secio "github.com/libp2p/go-libp2p-secio"
	yamux "github.com/libp2p/go-libp2p-yamux"
	lconfig "github.com/libp2p/go-libp2p/config"
	tcp "github.com/libp2p/go-tcp-transport"
	ws "github.com/libp2p/go-ws-transport"
)

func parseTransportOptions(opts []string) (lconfig.Option, error) {
	var lOpts []lconfig.Option

	for _, opt := range opts {
		switch strings.ToLower(opt) {
		case "tcp":
			lOpts = append(lOpts, libp2p.Transport(tcp.NewTCPTransport))

		case "ws":
			lOpts = append(lOpts, libp2p.Transport(ws.New))

		case "quic":
			lOpts = append(lOpts, libp2p.Transport(quic.NewTransport))

		/* note: utp has a broken gx dep
		case "utp":
			lOpts = append(lOpts, libp2p.Transport(utp.NewUtpTransport))
		*/

		/* note: WIP
		case "udp":
			lOpts = append(lOpts, libp2p.Transport(utp.NewUdpTransport))
		*/

		/* note: need to pass private key? But we didn't for quic...
		case "tls":
			lOpts = append(lOpts, libp2p.Transport(tls.New))
		*/

		case "none":
			if len(opts) > 1 {
				logger.Error("when using the 'none' transport option, cannot also specify other transport options")
				return nil, ErrImproperTransportOption
			}

			return libp2p.NoTransports, nil

		case "default":
			lOpts = append(lOpts, libp2p.DefaultTransports)

		default:
			logger.Errorf("unknown transport option: %s", opt)
			return nil, ErrUnknownTransportOption
		}
	}

	return libp2p.ChainOptions(lOpts...), nil
}

func parseMuxerOptions(opts [][]string) (lconfig.Option, error) {
	var lOpts []lconfig.Option

	for _, opt := range opts {
		if len(opt) != 2 {
			logger.Errorf("improper muxer format, expected ['name', 'type'], received %v", opt)
			return nil, ErrImproperMuxerOption
		}

		switch strings.ToLower(opt[0]) {
		case "yamux":
			lOpts = append(lOpts, libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport))

		case "mplex":
			lOpts = append(lOpts, libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport))

		// TODO: add others?
		default:
			logger.Errorf("unknown muxer option: %s", opt)
			return nil, ErrUnknownMuxerOption
		}
	}

	return libp2p.ChainOptions(lOpts...), nil
}

func parseSecurityOptions(opt string) (lconfig.Option, error) {
	switch strings.ToLower(opt) {
	case "secio":
		return libp2p.Security(secio.ID, secio.New), nil

	case "default":
		return libp2p.Security(secio.ID, secio.New), nil

	// TODO: add others?
	case "none":
		return libp2p.NoSecurity, nil

	default:
		logger.Errorf("unknown security option: %s", opt)
		return nil, ErrUnknownSecurityOption
	}
}

// note: it expects the peers to be in IPFS form
func connectToPeers(ctx context.Context, host host.Host, peers []string) error {
	for _, p := range peers {
		addr, err := ipfsaddr.ParseString(p)
		if err != nil {
			logger.Errorf("err parsing peer: %s\n%v", p, err)
			return err
		}

		pinfo, err := peerstore.InfoFromP2pAddr(addr.Multiaddr())
		if err != nil {
			logger.Errorf("err getting info from peerstore\n%v", err)
			return err
		}

		logger.Infof("full peer addr: %s", addr.String())
		logger.Infof("peer info: %v", pinfo)

		if err := host.Connect(ctx, *pinfo); err != nil {
			logger.Errorf("bootstrapping a peer failed\n%v", err)
			return err
		}

		logger.Infof("Connected to peer: %v", pinfo.ID)
	}

	return nil
}
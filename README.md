# X-Envoy
X-Envoy implements [Eden](https://arxiv.org/abs/2311.17454), SparkleX's parallel-verified messaging protocol. Built on a zero-knowledge MapReduce framework, Eden ensures ultra-fast, provably secure, and fully decentralized cross-chain communication between X-Chain and other blockchains.

Become an envoy in SparkleX by staking SparkleX tokens on X-Chain and running the X-Envoy service. As an envoy, you’ll support the omnichain liquidity network, maintain the system’s integrity, and earn rewards for your contributions! 😊

## How Envoy Works?
Envoys operate **independently**, monitoring activity on other chains and syncing changes to X-Chain without needing to communicate with each other. This design ensures strong resistance to collusion and exceptional scalability for the system.

Each envoy serves as a decentralized validator tasked with verifying committed messages from other chains. These messages generally include liquidity information updates or transaction requests that need to be synchronized with X-Chain. After receiving a message, an envoy privately verifies the message and votes the committed message using zero-knowledge proofs (ZK proofs) to ensure that the transaction is confirmed without exposing unnecessary details. Envoys then submit their votes to X-Chain, accompanied by a ZK proof.

## Components
X-Envoy has two components, `mapper` and `service`.
### Mapper
The `mapper` package is to:
- Generate ZK-VRF proofs
- Calculate votes for commited messages

### Service
The service manages and synchronizes Local Exit Roots (LERs) and Global Exit Roots (GERs) across various blockchain networks and X-Chain.
- Monitor configured other chains for LER updates
- Compare LERs between other chains and X-Chain
- Update X-Chain with new LERs when differences are detected

# Build & Run
1. Ensure the configuration file is correctly set up (default locations: `./config.yaml`, or `$HOME/.envoy/config.yaml`)
2. Run X-Envoy using the following command:
```
go run cmd/envoy/envoy.go
```

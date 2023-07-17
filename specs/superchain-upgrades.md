# Superchain Upgrades

Superchain upgrades, also known as forks or hardforks, implement consensus-breaking changes.

A Superchain upgrade requires the node software to support up to a given Protocol Version.
The version indicates support, the upgrade indicates the activation of new functionality.

This document lists the protocol versions of the OP-Stack, starting at the Bedrock upgrade,
as well as the default Superchain Targets.

Activation rule parameters of network upgrades are configured as part of the Superchain Target specification:
chains following the same Superchain Target upgrade synchronously.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Protocol Version](#protocol-version)
  - [Protocol Version Format](#protocol-version-format)
    - [Major versions](#major-versions)
    - [Minor versions](#minor-versions)
    - [Patch versions](#patch-versions)
  - [Protocol Version Exposure](#protocol-version-exposure)
- [Superchain Target](#superchain-target)
  - [Superchain Version signaling](#superchain-version-signaling)
- [Activation rules](#activation-rules)
  - [L2 Block-number based activation (deprecated)](#l2-block-number-based-activation-deprecated)
  - [L2 Block-timestamp based activation](#l2-block-timestamp-based-activation)
- [Post-Bedrock Network upgrades](#post-bedrock-network-upgrades)
  - [Regolith](#regolith)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Protocol Version

The Protocol Version documents the progression of the total set of canonical OP-Stack specifications.
Components of the OP-Stack implement the subset of their respective protocol component domain,
up to a given Protocol Version of the OP-Stack.

OP-Stack mods, i.e. non-canonical extensions to the OP-Stack, are not included in the versioning of the Protocol.
Instead, mods must specify which upstream Protocol Version they are based on and where breaking changes are made.
This ensures tooling of the OP-Stack can be shared and collaborated on with OP-Stack mods.

The Protocol Version is NOT a hardfork identifier, but rather indicates software-support for a well-defined set
of features introduced in past and future hardforks, not the activation of said hardforks.

Changes that can be included in prospective Protocol Versions may be included in the specifications as proposals,
with explicit notice of the Protocol Version they are based on.
This enables an iterative integration process into the canonical set of specifications,
but does not guarantee the proposed specifications become canonical.

Note that the Protocol Version only applies to the Protocol specifications with the Superchain Targets specified within.
This versioning is independent of the [Semver] versioning used in OP Stack smart-contracts,
and the [Semver]-versioned reference software of the OP-Stack.

### Protocol Version Format

The Protocol Version is [Semver]-compatible.
It is encoded as a single `uint256`:

```go
(major << (3 * 64)) | (minor << (2 * 64)) | (patch << 64) | 0
```

The version must be encoded as 32 bytes of `DATA` in JSON RPC usage.
The `major`, `minor`, and `patch` numbers are `uint64` bit numbers.
The lower 64 bits of the Protocol Version are reserved for future extensions.

[Semver]: https://semver.org/

#### Major versions

Major version changes indicate support for new consensus-breaking functionality.
Major versions should retain support for functionality of previous major versions for
syncing/indexing of historical chain data.
Implementations may drop support for previous Major versions, when there are viable alternatives,
e.g. `l2geth` for pre-Bedrock data.

Major Protocol Version changes require a governance vote to become canonical.

#### Minor versions

Minor version changes indicate support for backward compatible extensions,
including backward-compatible additions to the set of chains in a Superchain Target.
This may also include optional offchain functionality, such as additional syncing protocols.

Minor Protocol Version changes, as backward compatible, do not require a governance vote around protocol functionality.
Additions of chains to a Superchain Target are however subject to governance around the Superchain Target.

#### Patch versions

Patch version changes indicate backward compatible bug fixes and improvements.

### Protocol Version Exposure

The Protocol Version is not exposed to the application-layer environment:
hardforks already expose the change of functionality upon activation as required,
and the Protocol Version is meant for offchain usage only.
Again, indicating support rather than activation of functionality.
There is one exception however: signaling by onchain components to offchain components.
More about this in [Superchain Version signaling].

## Superchain Target

Changes to the L2 state-transition function are transitioned into deterministically across all nodes
through an **activation rule**.

Changes to L1 smart-contracts must be compatible with the latest activated L2 functionality,
and are executed through **L1 contract-upgrades**.

A Superchain Target defines a set of activation rules and L1 contract upgrades shared between OP-Stack chains,
to upgrade the chains collectively.

### Superchain Version signaling

Each Superchain Target tracks the protocol changes, and signals the recommended and required
Protocol Version ahead of activation of new breaking functionality.

Signaling is done through a L1 smart-contract that is monitored by the OP-Stack software.
Not all components of the OP-Stack are required to directly monitor L1 however:
cross-component APIs like the Engine API may be used to forward the Protocol Version signals,
to keep components encapsulated from L1.
See [`engine_signalOPStackVersionV1`](./exec-engine.md#enginesignalopstackversionv1).

## Activation rules

The below L2-block based activation rules may be applied in two contexts:

- The rollup node, specified through the rollup configuration (known as `rollup.json`),
  referencing L2 blocks (or block input-attributes) that pass through the derivation pipeline.
- The execution engine, specified through the chain configuration (known as the `config` part of `genesis.json`),
  referencing blocks or input-attributes that are part of, or applied to, the L2 chain.

For both types of configurations, some activation parameters may apply to all chains within the superchain,
and are then retrieved from the superchain target configuration.

### L2 Block-number based activation (deprecated)

Activation rule: `x != null && x >= upgradeNumber`

This block number based method has commonly been used in L1 up until the Bellatrix/Paris upgrade, a.k.a. The Merge,
which was upgraded through special rules.

This method is not superchain-compatible, as the activation-parameter is chain-specific
(different chains may have different block-heights at the same moment in time).

Starting at, and including, the L2 `block` with `block.number == x`, the upgrade rules apply.
If the upgrade block-number `x` is not specified in the configuration, the upgrade is ignored.

This applies to the L2 block number, not to the L1-origin block number.
This means that an L2 upgrade may be inactive, and then active, without changing the L1-origin.

### L2 Block-timestamp based activation

Activation rule: `x != null && x >= upgradeTime`

This is the preferred superchain upgrade activation-parameter type:
it is synchronous between all L2 chains and compatible with post-Merge timestamp-based chain upgrades in L1.

Starting at, and including, the L2 `block` with `block.timestamp == x`, the upgrade rules apply.
If the upgrade block-timestamp `x` is not specified in the configuration, the upgrade is ignored.

This applies to the L2 block timestamp, not to the L1-origin block timestamp.
This means that an L2 upgrade may be inactive, and then active, without changing the L1-origin.

This timestamp based method has become the default on L1 after the Bellatrix/Paris upgrade, a.k.a. The Merge,
because it can be planned in accordance with beacon-chain epochs and slots.

Note that the L2 version is not limited to timestamps that match L1 beacon-chain slots or epochs.
A timestamp may be chosen to be synchronous with a specific slot or epoch on L1,
but the matching L1-origin information may not be present at the time of activation on L2.

## Post-Bedrock Network upgrades

### Regolith

The Regolith upgrade, named after a material best described as "deposited dust on top of a layer of bedrock",
implements minor changes to deposit processing, based on reports of the Sherlock Audit-contest and findings in
the Bedrock Optimism Goerli testnet.

Summary of changes:

- The `isSystemTx` boolean is disabled, system transactions now use the same gas accounting rules as regular deposits.
- The actual deposit gas-usage is recorded in the receipt of the deposit transaction,
  and subtracted from the L2 block gas-pool.
  Unused gas of deposits is not refunded with ETH however, as it is burned on L1.
- The `nonce` value of the deposit sender account, before the transaction state-transition, is recorded in a new
  optional field (`depositNonce`), extending the transaction receipt (i.e. not present in pre-Regolith receipts).
- The recorded deposit `nonce` is used to correct the transaction and receipt metadata in RPC responses,
  including the `contractAddress` field of deposits that deploy contracts.
- The `gas` and `depositNonce` data is committed to as part of the consensus-representation of the receipt,
  enabling the data to be safely synced between independent L2 nodes.
- The L1-cost function was corrected to more closely match pre-Bedrock behavior.

The [deposit specification](./deposits.md) specifies the deposit changes of the Regolith upgrade in more detail.
The [execution engine specification](./exec-engine.md) specifies the L1 cost function difference.

The Regolith upgrade uses a *L2 block-timestamp* activation-rule, and is specified in both the
rollup-node (`regolith_time`) and execution engine (`config.regolithTime`).

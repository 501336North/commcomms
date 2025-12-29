# ADR-004: Hybrid DAO Governance

**Date:** 2025-12-29
**Status:** Accepted
**Deciders:** Product/Engineering Team

## Context

Community ownership is a key differentiator. Need:

1. **Member input on decisions** - Not just admin dictates
2. **Reputation-based influence** - Contributors should have more say
3. **Optional token layer** - For paid communities / treasuries
4. **Transparency** - Votes should be verifiable

Traditional platforms give all power to admins. DAOs give all power to token holders. We need a hybrid.

## Decision

Implement hybrid DAO governance:

### Voting Weight Calculation

```
vote_weight = reputation_weight + token_weight

Where:
  reputation_weight = user_reputation_in_community / total_community_reputation
  token_weight = user_token_balance / total_token_supply (if token deployed)
```

For communities without tokens: 100% reputation-weighted
For communities with tokens: Configurable split (default 50/50)

### Governance Flow

```
Member creates proposal
        ↓
Proposal enters voting period (24h-168h)
        ↓
Members cast votes (weighted)
        ↓
Voting ends
        ↓
Check quorum (default 20% participation)
        ↓
If quorum met → Result recorded on-chain
If not → Proposal expired
```

### On-Chain Recording (Base L2)

- Vote results recorded on Base L2 (low gas costs)
- Transaction hash stored in PostgreSQL
- Enables audit trail and dispute resolution
- Not required for proposal to take effect (async recording)

### Reputation Eligibility

| Action | Reputation Threshold |
|--------|---------------------|
| Vote on proposals | 0 (all members) |
| Create proposals | 100 (configurable) |
| Moderate content | 500 (admin-assigned) |

### No Forking Policy

- Knowledge belongs to community collectively
- Members can leave but cannot export wiki/summaries
- Prevents hostile forks while maintaining open contribution

## Consequences

### Positive
- Members feel ownership and invest more
- Reputation rewards contribution over tenure
- Token layer enables monetization without platform fee
- On-chain transparency for high-stakes decisions

### Negative
- Complexity for simple communities (overkill for small groups)
- On-chain recording adds latency and gas costs
- Reputation gaming potential (need anti-gaming measures)
- Token deployment has regulatory considerations

### Neutral
- Optional token layer - communities can opt out
- Quorum prevents low-participation decisions
- Admins can still take immediate action for moderation

## Alternatives Considered

### Alternative 1: Admin-Only Control (Traditional)
- **Pros**: Simple, fast decisions
- **Cons**: No member ownership, feels authoritarian
- **Rejected**: Core value is member ownership

### Alternative 2: Pure Token DAO
- **Pros**: True decentralization, clear incentives
- **Cons**: Plutocracy (rich = powerful), high friction
- **Rejected**: Reputation should matter, not just wealth

### Alternative 3: One Person One Vote
- **Pros**: Democratic, simple
- **Cons**: Sybil attack vulnerable, doesn't reward contribution
- **Rejected**: Want to incentivize valuable contributions

### Alternative 4: Quadratic Voting
- **Pros**: Reduces plutocracy, sophisticated
- **Cons**: Complex to explain, requires identity verification
- **Considered for v2**: May implement for high-stakes decisions

## Technical Implementation

### Smart Contract (Solidity)

```solidity
contract CommComsGovernance {
  function recordVote(
    bytes32 proposalId,
    bytes32 resultHash,
    uint256 totalWeight,
    address[] winners
  ) external onlyOracle;
}
```

### Off-Chain Voting

1. Votes collected off-chain (PostgreSQL)
2. At voting end, result hash computed
3. Oracle submits result to Base L2
4. Transaction hash stored back in PostgreSQL

## Related Decisions
- ADR-001: Hybrid Database Architecture (PostgreSQL for votes)
- ADR-002: AI-Native Knowledge System (no forking of knowledge)

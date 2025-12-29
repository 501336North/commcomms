/**
 * CommComms API Types
 * Generated from OpenAPI spec
 */

// ============================================
// AUTH TYPES
// ============================================

export interface RegisterRequest {
  email: string;
  password: string;
  handle: string;
  inviteCode: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RefreshRequest {
  refreshToken: string;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: UserProfile;
}

// ============================================
// USER TYPES
// ============================================

export interface UserProfile {
  id: string;
  handle: string;
  displayName?: string;
  bio?: string;
  avatarUrl?: string;
  reputation: number;
  createdAt: string;
  currentLocation?: Location;
}

export interface UpdateUserRequest {
  displayName?: string;
  bio?: string;
  avatarUrl?: string;
}

export interface ReputationDetails {
  total: number;
  rank: number;
  breakdown: {
    messages: number;
    helpfulMarks: number;
    threadsStarted: number;
    votesParticipated: number;
  };
}

// ============================================
// COMMUNITY TYPES
// ============================================

export interface Community {
  id: string;
  name: string;
  description?: string;
  isPrivate: boolean;
  memberCount: number;
  settings: CommunitySettings;
  createdAt: string;
}

export interface CommunitySettings {
  echoEnabled: boolean;
  echoTtlHours: number;
  echoConfidenceThreshold: number;
  minReputationToPropose: number;
  votingQuorumPercent: number;
}

export interface CommunityList {
  data: Community[];
}

export interface CreateCommunityRequest {
  name: string;
  description?: string;
  isPrivate?: boolean;
}

export interface UpdateCommunityRequest {
  name?: string;
  description?: string;
  settings?: Partial<CommunitySettings>;
}

export interface Invite {
  code: string;
  url: string;
  expiresAt: string;
  maxUses?: number;
  usesRemaining?: number;
}

export interface CreateInviteRequest {
  expiresInDays?: number;
  maxUses?: number;
}

export type MemberRole = 'owner' | 'admin' | 'moderator' | 'member';

export interface Member {
  user: UserProfile;
  role: MemberRole;
  joinedAt: string;
  reputationInCommunity: number;
}

export interface MemberList {
  data: Member[];
  meta: PaginationMeta;
}

// ============================================
// CHANNEL TYPES
// ============================================

export interface Channel {
  id: string;
  name: string;
  description?: string;
  threadCount: number;
  lastActivityAt: string;
}

export interface ChannelList {
  data: Channel[];
}

export interface CreateChannelRequest {
  name: string;
  description?: string;
}

// ============================================
// THREAD TYPES
// ============================================

export interface Thread {
  id: string;
  title: string;
  author: UserProfile;
  messageCount: number;
  participantCount: number;
  hasSummary: boolean;
  createdAt: string;
  lastMessageAt: string;
}

export interface ThreadList {
  data: Thread[];
  meta: PaginationMeta;
}

export interface ThreadWithMessages {
  thread: Thread;
  summary?: ThreadSummary;
  messages: Message[];
  meta: PaginationMeta;
}

export interface CreateThreadRequest {
  title: string;
  initialMessage?: string;
}

// ============================================
// MESSAGE TYPES
// ============================================

export interface Message {
  id: string;
  content: string;
  author: UserProfile;
  isEcho: boolean;
  expiresAt?: string;
  reactions: Reaction[];
  entities: EntityRef[];
  createdAt: string;
  editedAt?: string;
}

export interface SendMessageRequest {
  content: string;
}

export interface EditMessageRequest {
  content: string;
}

export interface Reaction {
  emoji: string;
  count: number;
  users: string[];
}

export interface ReactionRequest {
  emoji: string;
}

// ============================================
// KNOWLEDGE TYPES
// ============================================

export type EntityType = 'location' | 'topic' | 'person';

export interface Entity {
  id: string;
  type: EntityType;
  name: string;
  mentionCount: number;
  lastMentionedAt: string;
}

export interface EntityRef {
  id: string;
  type: EntityType;
  name: string;
}

export interface EntityList {
  data: Entity[];
}

export interface EntityWithCard {
  entity: Entity;
  card: KnowledgeCard;
  relatedThreads: Thread[];
}

export interface KnowledgeCard {
  summary: string;
  keyPoints: string[];
  contributors: UserProfile[];
  lastUpdatedAt: string;
}

export interface ThreadSummary {
  summary: string;
  keyPoints: string[];
  entities: EntityRef[];
  generatedAt: string;
}

// ============================================
// SEARCH TYPES
// ============================================

export type SearchType = 'keyword' | 'semantic' | 'hybrid';
export type SearchResultType = 'message' | 'thread' | 'entity';

export interface SearchResults {
  data: SearchResult[];
  meta: PaginationMeta;
}

export interface SearchResult {
  type: SearchResultType;
  score: number;
  message?: Message;
  thread?: Thread;
  entity?: Entity;
  highlight?: string;
}

// ============================================
// GOVERNANCE TYPES
// ============================================

export type ProposalStatus = 'active' | 'passed' | 'rejected' | 'expired';

export interface Proposal {
  id: string;
  title: string;
  description: string;
  author: UserProfile;
  status: ProposalStatus;
  options: ProposalOption[];
  votingEndsAt: string;
  createdAt: string;
}

export interface ProposalOption {
  id: string;
  label: string;
  voteWeight: number;
  voteCount: number;
}

export interface ProposalList {
  data: Proposal[];
}

export interface ProposalDetails {
  proposal: Proposal;
  quorumReached: boolean;
  totalVoteWeight: number;
  onChainTxHash?: string;
}

export interface CreateProposalRequest {
  title: string;
  description: string;
  options: string[];
  votingDurationHours: number;
}

export interface CastVoteRequest {
  optionId: string;
}

export interface VoteReceipt {
  proposalId: string;
  optionId: string;
  weight: number;
  txHash?: string;
}

// ============================================
// TOKEN TYPES
// ============================================

export interface WalletDetails {
  address: string;
  chainId: number;
  connectedAt: string;
}

export interface ConnectWalletRequest {
  address: string;
  signature: string;
  message: string;
}

export interface TokenInfo {
  address: string;
  name: string;
  symbol: string;
  totalSupply: string;
  decimals: number;
  chainId: number;
  deployedAt: string;
}

export interface DeployTokenRequest {
  name: string;
  symbol: string;
  initialSupply: string;
}

export type TokenDeploymentStatusType = 'pending' | 'confirmed' | 'failed';

export interface TokenDeploymentStatus {
  status: TokenDeploymentStatusType;
  txHash?: string;
  tokenAddress?: string;
}

export interface DistributeTokensRequest {
  distributions: Array<{
    userId: string;
    amount: string;
  }>;
}

// ============================================
// LOCATION TYPES
// ============================================

export interface Location {
  city?: string;
  country: string;
  countryCode: string;
  latitude?: number;
  longitude?: number;
}

export type LocationVisibility = 'precise' | 'city' | 'country';

export interface CheckIn {
  id: string;
  location: Location;
  visibility: LocationVisibility;
  checkedInAt: string;
}

export interface CheckInRequest {
  latitude: number;
  longitude: number;
  visibility?: LocationVisibility;
}

export interface NearbyMembers {
  location: string;
  members: Array<{
    user: UserProfile;
    checkedInAt: string;
    durationDays: number;
  }>;
}

// ============================================
// PRESENCE TYPES
// ============================================

export interface PresenceList {
  online: Array<{
    user: UserProfile;
    lastSeenAt: string;
  }>;
}

// ============================================
// COMMON TYPES
// ============================================

export interface ErrorResponse {
  error: string;
  code?: string;
  details?: Record<string, unknown>;
}

export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export type AsyncJobStatus = 'pending' | 'processing' | 'completed' | 'failed';

export interface AsyncJobResponse {
  jobId: string;
  status: AsyncJobStatus;
  estimatedCompletionSeconds?: number;
}

// ============================================
// API CLIENT INTERFACES (for mocking in tests)
// ============================================

export interface AuthAPI {
  register(data: RegisterRequest): Promise<AuthResponse>;
  login(data: LoginRequest): Promise<AuthResponse>;
  logout(): Promise<void>;
  refresh(data: RefreshRequest): Promise<AuthResponse>;
}

export interface UserAPI {
  getCurrentUser(): Promise<UserProfile>;
  updateCurrentUser(data: UpdateUserRequest): Promise<UserProfile>;
  getUser(userId: string): Promise<UserProfile>;
  getUserReputation(userId: string): Promise<ReputationDetails>;
}

export interface CommunityAPI {
  listCommunities(): Promise<CommunityList>;
  createCommunity(data: CreateCommunityRequest): Promise<Community>;
  getCommunity(communityId: string): Promise<Community>;
  updateCommunity(communityId: string, data: UpdateCommunityRequest): Promise<Community>;
  createInvite(communityId: string, data: CreateInviteRequest): Promise<Invite>;
  listMembers(communityId: string, page?: number, limit?: number): Promise<MemberList>;
}

export interface ChannelAPI {
  listChannels(communityId: string): Promise<ChannelList>;
  createChannel(communityId: string, data: CreateChannelRequest): Promise<Channel>;
}

export interface ThreadAPI {
  listThreads(channelId: string, page?: number, limit?: number): Promise<ThreadList>;
  createThread(channelId: string, data: CreateThreadRequest): Promise<Thread>;
  getThread(threadId: string, page?: number, limit?: number): Promise<ThreadWithMessages>;
}

export interface MessageAPI {
  sendMessage(threadId: string, data: SendMessageRequest): Promise<Message>;
  editMessage(messageId: string, data: EditMessageRequest): Promise<Message>;
  deleteMessage(messageId: string): Promise<void>;
  addReaction(messageId: string, data: ReactionRequest): Promise<void>;
}

export interface KnowledgeAPI {
  listEntities(communityId: string, type?: EntityType): Promise<EntityList>;
  getEntity(entityId: string): Promise<EntityWithCard>;
  getThreadSummary(threadId: string): Promise<ThreadSummary>;
  generateSummary(threadId: string): Promise<AsyncJobResponse>;
}

export interface SearchAPI {
  search(
    communityId: string,
    query: string,
    options?: {
      type?: SearchType;
      entityFilter?: string;
      page?: number;
      limit?: number;
    }
  ): Promise<SearchResults>;
}

export interface GovernanceAPI {
  listProposals(communityId: string, status?: ProposalStatus): Promise<ProposalList>;
  createProposal(communityId: string, data: CreateProposalRequest): Promise<Proposal>;
  getProposal(proposalId: string): Promise<ProposalDetails>;
  castVote(proposalId: string, data: CastVoteRequest): Promise<VoteReceipt>;
  changeVote(proposalId: string, data: CastVoteRequest): Promise<VoteReceipt>;
}

export interface TokenAPI {
  getWallet(): Promise<WalletDetails>;
  connectWallet(data: ConnectWalletRequest): Promise<WalletDetails>;
  disconnectWallet(): Promise<void>;
  getCommunityToken(communityId: string): Promise<TokenInfo>;
  deployToken(communityId: string, data: DeployTokenRequest): Promise<TokenDeploymentStatus>;
  distributeTokens(communityId: string, data: DistributeTokensRequest): Promise<AsyncJobResponse>;
}

export interface LocationAPI {
  checkIn(data: CheckInRequest): Promise<CheckIn>;
  getCurrentCheckIn(): Promise<CheckIn>;
  findNearby(
    communityId: string,
    location: string,
    radius?: 'city' | 'country' | 'region'
  ): Promise<NearbyMembers>;
}

export interface PresenceAPI {
  getPresence(communityId: string): Promise<PresenceList>;
}

// ============================================
// WEBSOCKET EVENTS
// ============================================

export type WebSocketEventType =
  | 'message:new'
  | 'message:edit'
  | 'message:delete'
  | 'message:reaction'
  | 'presence:online'
  | 'presence:offline'
  | 'presence:typing'
  | 'thread:summary'
  | 'echo:response';

export interface WebSocketEvent<T = unknown> {
  type: WebSocketEventType;
  payload: T;
  timestamp: string;
}

export interface MessageNewEvent {
  threadId: string;
  message: Message;
}

export interface MessageEditEvent {
  threadId: string;
  messageId: string;
  content: string;
  editedAt: string;
}

export interface MessageDeleteEvent {
  threadId: string;
  messageId: string;
}

export interface PresenceEvent {
  userId: string;
  handle: string;
  communityId: string;
}

export interface TypingEvent {
  threadId: string;
  userId: string;
  handle: string;
}

export interface EchoResponseEvent {
  threadId: string;
  message: Message;
  sourceThreads: Array<{ id: string; title: string }>;
}

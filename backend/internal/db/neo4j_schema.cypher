// CommComms Knowledge Graph Schema
// Neo4j 5.x
//
// This schema stores the knowledge graph - entities extracted from conversations
// and their relationships. Linked to PostgreSQL via entity IDs.

// ============================================
// CONSTRAINTS (unique identifiers)
// ============================================

// Entity constraints
CREATE CONSTRAINT entity_id IF NOT EXISTS
FOR (e:Entity) REQUIRE e.id IS UNIQUE;

CREATE CONSTRAINT location_name IF NOT EXISTS
FOR (l:Location) REQUIRE l.name IS UNIQUE;

CREATE CONSTRAINT topic_name IF NOT EXISTS
FOR (t:Topic) REQUIRE t.name IS UNIQUE;

// Thread/Message references (link to PostgreSQL)
CREATE CONSTRAINT thread_ref IF NOT EXISTS
FOR (t:ThreadRef) REQUIRE t.pgId IS UNIQUE;

CREATE CONSTRAINT message_ref IF NOT EXISTS
FOR (m:MessageRef) REQUIRE m.pgId IS UNIQUE;

CREATE CONSTRAINT community_ref IF NOT EXISTS
FOR (c:CommunityRef) REQUIRE c.pgId IS UNIQUE;

CREATE CONSTRAINT user_ref IF NOT EXISTS
FOR (u:UserRef) REQUIRE u.pgId IS UNIQUE;

// ============================================
// INDEXES
// ============================================

// Full-text search on entity names
CREATE FULLTEXT INDEX entity_name_fulltext IF NOT EXISTS
FOR (e:Entity) ON EACH [e.name, e.aliases];

// For finding entities by type
CREATE INDEX entity_type IF NOT EXISTS
FOR (e:Entity) ON (e.type);

// For finding recent mentions
CREATE INDEX mention_timestamp IF NOT EXISTS
FOR ()-[m:MENTIONS]-() ON (m.timestamp);

// For finding summaries by community
CREATE INDEX summary_community IF NOT EXISTS
FOR (s:Summary) ON (s.communityId);

// ============================================
// NODE TYPES
// ============================================

// Entity - base label for all knowledge entities
// Properties:
//   id: UUID (unique identifier)
//   type: 'location' | 'topic' | 'person'
//   name: string (canonical name)
//   aliases: string[] (alternative names)
//   mentionCount: int (total mentions)
//   lastMentionedAt: datetime
//   createdAt: datetime

// Location - places (cities, countries, venues)
// Additional properties:
//   city: string
//   country: string
//   countryCode: string
//   latitude: float
//   longitude: float

// Topic - discussion topics (coworking, visa, etc.)
// Additional properties:
//   category: string (optional grouping)

// Person - mentioned people (not platform users)
// Additional properties:
//   description: string

// Summary - thread summary attached to entities
// Properties:
//   id: UUID
//   threadId: UUID (PostgreSQL reference)
//   communityId: UUID
//   content: string
//   keyPoints: string[]
//   generatedAt: datetime

// ThreadRef - reference to PostgreSQL thread
// Properties:
//   pgId: UUID (PostgreSQL id)
//   title: string (cached for display)

// MessageRef - reference to PostgreSQL message
// Properties:
//   pgId: UUID (PostgreSQL id)

// UserRef - reference to PostgreSQL user
// Properties:
//   pgId: UUID
//   handle: string (cached)

// CommunityRef - reference to PostgreSQL community
// Properties:
//   pgId: UUID
//   name: string (cached)

// ============================================
// RELATIONSHIP TYPES
// ============================================

// MENTIONS - an entity was mentioned in a message
// (MessageRef)-[:MENTIONS {confidence: float, timestamp: datetime}]->(Entity)

// SUMMARIZES - a summary covers entities
// (Summary)-[:SUMMARIZES {relevance: float}]->(Entity)

// FROM_THREAD - summary belongs to thread
// (Summary)-[:FROM_THREAD]->(ThreadRef)

// CONTAINS - thread contains message
// (ThreadRef)-[:CONTAINS]->(MessageRef)

// AUTHORED_BY - message/thread authored by user
// (MessageRef)-[:AUTHORED_BY]->(UserRef)
// (ThreadRef)-[:AUTHORED_BY]->(UserRef)

// BELONGS_TO - entity belongs to community
// (Entity)-[:BELONGS_TO]->(CommunityRef)

// RELATED_TO - entities related to each other
// (Entity)-[:RELATED_TO {strength: float, type: string}]->(Entity)

// IN_LOCATION - sub-location relationship
// (Location)-[:IN_LOCATION]->(Location)
// e.g., (Lisbon)-[:IN_LOCATION]->(Portugal)

// ABOUT - topic hierarchy
// (Topic)-[:ABOUT]->(Topic)
// e.g., (Coworking Spaces)-[:ABOUT]->(Remote Work)

// ============================================
// EXAMPLE QUERIES
// ============================================

// Find all knowledge about Lisbon
// MATCH (l:Location {name: 'Lisbon'})<-[:MENTIONS]-(m:MessageRef)
// MATCH (m)<-[:CONTAINS]-(t:ThreadRef)
// RETURN l, m, t ORDER BY m.timestamp DESC LIMIT 50

// Find related entities
// MATCH (l:Location {name: 'Lisbon'})-[:RELATED_TO*1..2]-(related:Entity)
// RETURN related.name, related.type, COUNT(*) as connections
// ORDER BY connections DESC

// Semantic search (combined with pgvector results)
// Find entities mentioned in threads similar to query
// MATCH (e:Entity)<-[:MENTIONS]-(m:MessageRef)
// WHERE m.pgId IN $messageIds  // From pgvector similarity search
// RETURN e.name, e.type, COUNT(*) as mentions
// ORDER BY mentions DESC

// Get knowledge card for entity
// MATCH (e:Entity {id: $entityId})
// OPTIONAL MATCH (s:Summary)-[:SUMMARIZES]->(e)
// OPTIONAL MATCH (e)<-[:MENTIONS]-(m:MessageRef)<-[:CONTAINS]-(t:ThreadRef)
// WITH e, s, t ORDER BY s.generatedAt DESC, m.timestamp DESC
// RETURN e, COLLECT(DISTINCT s)[0..5] as summaries, COLLECT(DISTINCT t)[0..10] as threads

// Find users who contributed to entity knowledge
// MATCH (e:Entity {id: $entityId})<-[:MENTIONS]-(m:MessageRef)-[:AUTHORED_BY]->(u:UserRef)
// RETURN u.handle, u.pgId, COUNT(*) as contributions
// ORDER BY contributions DESC

// Community knowledge graph stats
// MATCH (c:CommunityRef {pgId: $communityId})<-[:BELONGS_TO]-(e:Entity)
// RETURN e.type, COUNT(*) as count
// ORDER BY count DESC

// ============================================
// INITIALIZATION QUERIES
// ============================================

// Create common location hierarchy
// MERGE (portugal:Location {name: 'Portugal', type: 'location'})
// SET portugal.country = 'Portugal', portugal.countryCode = 'PT'
// MERGE (lisbon:Location {name: 'Lisbon', type: 'location'})
// SET lisbon.city = 'Lisbon', lisbon.country = 'Portugal', lisbon.countryCode = 'PT'
// MERGE (lisbon)-[:IN_LOCATION]->(portugal)

// Create common topic categories
// MERGE (remote:Topic {name: 'Remote Work', type: 'topic'})
// MERGE (coworking:Topic {name: 'Coworking', type: 'topic'})
// MERGE (coworking)-[:ABOUT]->(remote)
// MERGE (visa:Topic {name: 'Visa', type: 'topic'})
// MERGE (travel:Topic {name: 'Travel', type: 'topic'})
// MERGE (visa)-[:ABOUT]->(travel)

// ============================================
// MAINTENANCE QUERIES
// ============================================

// Update mention counts (run periodically)
// MATCH (e:Entity)
// SET e.mentionCount = SIZE([(e)<-[:MENTIONS]-() | 1])

// Find orphaned entities (no recent mentions)
// MATCH (e:Entity)
// WHERE e.lastMentionedAt < datetime() - duration('P90D')
// AND e.mentionCount < 5
// RETURN e.name, e.type, e.lastMentionedAt

// Merge duplicate entities (manual review)
// MATCH (e1:Entity {name: 'Lisbon'}), (e2:Entity {name: 'Lisboa'})
// CALL apoc.refactor.mergeNodes([e1, e2], {
//   properties: 'combine',
//   mergeRels: true
// })
// YIELD node
// RETURN node

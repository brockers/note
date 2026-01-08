# Offline-First Sync Patterns for TypeScript Mobile Applications (2024-2025)

**Comprehensive guide for building robust offline-first mobile applications with reliable synchronization patterns, conflict resolution, and enterprise-grade data consistency.**

---

## Offline-First Architecture Principles

### Local-First Software Philosophy

**Core Principle**: "Your data should be yours, and you should be able to access and modify it regardless of network connectivity."

**Key Tenets**:
1. **Primary data lives on user's device** - Local operations are always fast and available
2. **Network is treated as an enhancement** - Online connectivity provides sync, not core functionality  
3. **User maintains agency** - Can use app fully without internet, data ownership stays with user
4. **Multi-device sync without central coordination** - Devices can sync peer-to-peer when needed
5. **Collaborative capabilities built-in** - Multiple users can work on same data with conflict resolution

### Offline-First vs. Offline-Capable

```typescript
// ❌ Offline-Capable (Network-first with caching)
class NetworkFirstService {
  async getData(id: string) {
    try {
      // Try network first
      const data = await this.api.fetch(id);
      await this.cache.store(id, data);
      return data;
    } catch (error) {
      // Fall back to cache
      return await this.cache.get(id);
    }
  }
}

// ✅ Offline-First (Local-first with sync)
class OfflineFirstService {
  async getData(id: string) {
    // Always serve from local database first
    const localData = await this.localDb.get(id);
    
    // Trigger background sync if needed
    if (this.isOnline && this.needsSync(localData)) {
      this.backgroundSync.queue({ type: 'fetch', id });
    }
    
    return localData;
  }
}
```

---

## Database Solutions for Offline-First Apps

### WatermelonDB (Recommended for React Native)

**Why WatermelonDB**:
- Built specifically for React Native offline-first apps
- Lazy loading and reactive queries for performance
- SQLite-based with multi-threading support
- Excellent TypeScript integration
- Handles thousands of records efficiently

**Schema Setup**:
```typescript
import { appSchema, tableSchema } from '@nozbe/watermelondb';

export const schema = appSchema({
  version: 1,
  tables: [
    tableSchema({
      name: 'photos',
      columns: [
        { name: 'local_path', type: 'string' },
        { name: 'remote_url', type: 'string', isOptional: true },
        { name: 'department_id', type: 'string' },
        { name: 'project_id', type: 'string' },
        { name: 'captured_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
        { name: 'sync_status', type: 'string' },
        { name: 'metadata_json', type: 'string' },
        { name: 'conflict_version', type: 'number', isOptional: true },
        { name: 'last_modified', type: 'number' }
      ]
    }),
    tableSchema({
      name: 'sync_queue',
      columns: [
        { name: 'operation_type', type: 'string' }, // 'create', 'update', 'delete'
        { name: 'table_name', type: 'string' },
        { name: 'record_id', type: 'string' },
        { name: 'changes_json', type: 'string' },
        { name: 'attempts', type: 'number' },
        { name: 'last_attempt', type: 'number', isOptional: true },
        { name: 'created_at', type: 'number' }
      ]
    }),
    tableSchema({
      name: 'conflict_log',
      columns: [
        { name: 'table_name', type: 'string' },
        { name: 'record_id', type: 'string' },
        { name: 'local_version_json', type: 'string' },
        { name: 'remote_version_json', type: 'string' },
        { name: 'resolution_strategy', type: 'string' },
        { name: 'resolved_at', type: 'number', isOptional: true },
        { name: 'created_at', type: 'number' }
      ]
    })
  ]
});
```

**Model Implementation with Sync Support**:
```typescript
import { Model, field, date, readonly } from '@nozbe/watermelondb/decorators';
import { SyncStatus, SyncableRecord } from '../types/sync.types';

export class Photo extends Model implements SyncableRecord {
  static table = 'photos';

  @field('local_path') localPath!: string;
  @field('remote_url') remoteUrl?: string;
  @field('department_id') departmentId!: string;
  @field('project_id') projectId!: string;
  @date('captured_at') capturedAt!: Date;
  @date('synced_at') syncedAt?: Date;
  @field('sync_status') syncStatus!: SyncStatus;
  @field('metadata_json') private metadataJson!: string;
  @field('conflict_version') conflictVersion?: number;
  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @field('last_modified') lastModified!: number;

  get metadata(): PhotoMetadata {
    return JSON.parse(this.metadataJson);
  }

  set metadata(value: PhotoMetadata) {
    this.metadataJson = JSON.stringify(value);
  }

  get needsSync(): boolean {
    return ['pending', 'failed'].includes(this.syncStatus);
  }

  get isConflicted(): boolean {
    return this.syncStatus === 'conflicted';
  }

  // Optimistic update method
  async updateOptimistically(changes: Partial<PhotoMetadata>): Promise<void> {
    await this.update(photo => {
      photo.metadata = { ...photo.metadata, ...changes };
      photo.syncStatus = 'pending';
      photo.lastModified = Date.now();
    });
  }
}

export interface SyncableRecord {
  syncStatus: SyncStatus;
  lastModified: number;
  needsSync: boolean;
  isConflicted: boolean;
}

export type SyncStatus = 
  | 'synced'     // Successfully synchronized
  | 'pending'    // Waiting for sync
  | 'syncing'    // Currently being synced
  | 'failed'     // Sync failed
  | 'conflicted' // Has unresolved conflicts;
```

### PowerSync for MongoDB Integration

**PowerSync Advantages**:
- Real-time sync with MongoDB using change streams
- Built-in conflict resolution
- Handles complex relational data
- Enterprise-ready with high availability

**Configuration**:
```typescript
import { PowerSyncDatabase } from '@powersync/react-native';
import { Column, ColumnType, Schema, Table } from '@powersync/common';

const schema = new Schema([
  new Table({
    name: 'photos',
    columns: [
      new Column({ name: 'local_path', type: ColumnType.text }),
      new Column({ name: 'remote_url', type: ColumnType.text }),
      new Column({ name: 'department_id', type: ColumnType.text }),
      new Column({ name: 'project_id', type: ColumnType.text }),
      new Column({ name: 'captured_at', type: ColumnType.integer }),
      new Column({ name: 'sync_status', type: ColumnType.text }),
      new Column({ name: 'server_created_at', type: ColumnType.text })
    ]
  }),
  new Table({
    name: 'departments',
    columns: [
      new Column({ name: 'name', type: ColumnType.text }),
      new Column({ name: 'code', type: ColumnType.text }),
      new Column({ name: 'active', type: ColumnType.integer })
    ]
  })
]);

export class PowerSyncService {
  private db: PowerSyncDatabase;
  private uploadQueue = new Map<string, UploadTask>();

  constructor() {
    this.db = new PowerSyncDatabase({
      schema,
      database: { dbFilename: 'manufacturing.sqlite' }
    });
  }

  async initialize(): Promise<void> {
    await this.db.init();
    
    // Connect to PowerSync service
    await this.db.connect({
      powerSyncUrl: 'https://your-powersync-instance.com',
      token: await this.getAuthToken(),
      params: {
        department_id: await this.getCurrentDepartmentId()
      }
    });

    this.startUploadQueueProcessor();
  }

  private async startUploadQueueProcessor(): Promise<void> {
    // Process upload queue for large files (photos)
    setInterval(async () => {
      if (this.uploadQueue.size === 0) return;

      const tasks = Array.from(this.uploadQueue.values())
        .filter(task => task.status === 'pending')
        .slice(0, 3); // Process 3 uploads concurrently

      for (const task of tasks) {
        this.processUploadTask(task);
      }
    }, 5000); // Check every 5 seconds
  }
}
```

### RxDB for Reactive Applications

**RxDB Benefits**:
- Reactive queries with RxJS observables
- Multi-tab synchronization
- Plugin ecosystem for encryption, replication
- Works with React, Angular, Vue

**Setup Example**:
```typescript
import { createRxDatabase, addRxPlugin } from 'rxdb';
import { getRxStorageDexie } from 'rxdb/plugins/storage-dexie';
import { RxDBReplicationCouchDBPlugin } from 'rxdb/plugins/replication-couchdb';

addRxPlugin(RxDBReplicationCouchDBPlugin);

const photoSchema = {
  title: 'photo schema',
  version: 0,
  primaryKey: 'id',
  type: 'object',
  properties: {
    id: { type: 'string', maxLength: 100 },
    localPath: { type: 'string' },
    remoteUrl: { type: 'string' },
    departmentId: { type: 'string' },
    projectId: { type: 'string' },
    capturedAt: { type: 'string', format: 'date-time' },
    syncStatus: { type: 'string', enum: ['pending', 'synced', 'failed'] },
    metadata: { type: 'object' }
  },
  required: ['id', 'localPath', 'departmentId', 'projectId', 'capturedAt']
};

export class RxDBOfflineService {
  private database: RxDatabase;

  async initialize() {
    this.database = await createRxDatabase({
      name: 'manufacturing_db',
      storage: getRxStorageDexie()
    });

    const photosCollection = await this.database.addCollections({
      photos: {
        schema: photoSchema
      }
    });

    // Set up replication
    const replicationState = photosCollection.photos.syncCouchDB({
      remote: 'http://localhost:5984/manufacturing/',
      waitForLeadership: false,
      direction: {
        pull: true,
        push: true
      },
      options: {
        live: true,
        retry: true
      }
    });

    return this.database;
  }
}
```

---

## Synchronization Patterns

### Event Sourcing for Audit Trails

**Perfect for Manufacturing/Compliance Use Cases**:
```typescript
interface ManufacturingEvent {
  id: string;
  aggregateId: string; // Photo ID or inspection session ID
  eventType: string;
  eventData: Record<string, unknown>;
  timestamp: number;
  userId: string;
  deviceId: string;
  departmentId: string;
  sequenceNumber: number;
  checksum: string; // For data integrity
}

export class EventStore {
  constructor(private localDb: WatermelonDatabase) {}

  async appendEvent(event: Omit<ManufacturingEvent, 'id' | 'timestamp' | 'sequenceNumber' | 'checksum'>): Promise<void> {
    const sequenceNumber = await this.getNextSequenceNumber(event.aggregateId);
    
    const manufacturingEvent: ManufacturingEvent = {
      ...event,
      id: this.generateEventId(),
      timestamp: Date.now(),
      sequenceNumber,
      checksum: this.calculateChecksum(event)
    };

    // Store locally first (always succeeds)
    await this.localDb.write(async () => {
      await this.localDb.get('events').create(eventRecord => {
        Object.assign(eventRecord, manufacturingEvent);
      });
    });

    // Queue for sync
    await this.queueEventForSync(manufacturingEvent);
  }

  async getEventHistory(aggregateId: string): Promise<ManufacturingEvent[]> {
    const events = await this.localDb.get('events')
      .query(
        Q.where('aggregate_id', aggregateId),
        Q.sortBy('sequence_number', Q.asc)
      )
      .fetch();

    return events.map(event => this.eventFromRecord(event));
  }

  // Reconstruct current state from events
  async reconstructAggregate<T>(aggregateId: string, initialState: T): Promise<T> {
    const events = await this.getEventHistory(aggregateId);
    
    return events.reduce((state, event) => {
      return this.applyEvent(state, event);
    }, initialState);
  }

  private applyEvent<T>(state: T, event: ManufacturingEvent): T {
    switch (event.eventType) {
      case 'photo_captured':
        return this.applyPhotoCaptured(state, event.eventData);
      case 'quality_inspection_started':
        return this.applyQualityInspectionStarted(state, event.eventData);
      case 'metadata_updated':
        return this.applyMetadataUpdated(state, event.eventData);
      default:
        console.warn(`Unknown event type: ${event.eventType}`);
        return state;
    }
  }

  private calculateChecksum(event: Partial<ManufacturingEvent>): string {
    const { eventType, eventData, userId, deviceId } = event;
    const payload = JSON.stringify({ eventType, eventData, userId, deviceId });
    
    // Simple checksum - in production use crypto hash
    return Buffer.from(payload).toString('base64');
  }
}
```

### CRDT (Conflict-free Replicated Data Types)

**For Automatic Conflict Resolution**:
```typescript
import { LWWRegister, GSet, ORMap } from 'crdt-js';

interface PhotoMetadataCRDT {
  basicInfo: LWWRegister<BasicPhotoInfo>; // Last-Writer-Wins for simple fields
  tags: GSet<string>; // Grow-only set for tags
  inspectionData: ORMap<string, LWWRegister<unknown>>; // Map of inspection fields
}

export class CRDTSyncManager {
  private crdtStates = new Map<string, PhotoMetadataCRDT>();

  updatePhotoMetadata(photoId: string, field: string, value: unknown, timestamp?: number): void {
    let crdt = this.crdtStates.get(photoId);
    
    if (!crdt) {
      crdt = {
        basicInfo: new LWWRegister(),
        tags: new GSet(),
        inspectionData: new ORMap()
      };
      this.crdtStates.set(photoId, crdt);
    }

    const ts = timestamp || Date.now();

    if (field === 'tags' && Array.isArray(value)) {
      // Tags are append-only for audit trail
      value.forEach(tag => crdt!.tags.add(tag));
    } else if (field.startsWith('inspection.')) {
      // Inspection data uses Last-Writer-Wins
      const inspectionField = field.substring(11);
      const register = crdt.inspectionData.get(inspectionField) || new LWWRegister();
      register.set(value, ts);
      crdt.inspectionData.set(inspectionField, register);
    } else {
      // Basic info uses Last-Writer-Wins
      const currentBasic = crdt.basicInfo.value() || {};
      crdt.basicInfo.set({ ...currentBasic, [field]: value }, ts);
    }
  }

  mergeFromRemote(photoId: string, remoteCRDT: PhotoMetadataCRDT): PhotoMetadataCRDT {
    const localCRDT = this.crdtStates.get(photoId);
    
    if (!localCRDT) {
      this.crdtStates.set(photoId, remoteCRDT);
      return remoteCRDT;
    }

    // Merge CRDTs - this is conflict-free by mathematical properties
    localCRDT.basicInfo.merge(remoteCRDT.basicInfo);
    localCRDT.tags.merge(remoteCRDT.tags);
    localCRDT.inspectionData.merge(remoteCRDT.inspectionData);

    return localCRDT;
  }

  getCurrentState(photoId: string): PhotoMetadata | null {
    const crdt = this.crdtStates.get(photoId);
    if (!crdt) return null;

    // Convert CRDT state to regular PhotoMetadata
    const basicInfo = crdt.basicInfo.value() || {};
    const tags = Array.from(crdt.tags.values());
    const inspectionData: Record<string, unknown> = {};

    crdt.inspectionData.keys().forEach(key => {
      const register = crdt.inspectionData.get(key);
      if (register) {
        inspectionData[key] = register.value();
      }
    });

    return {
      ...basicInfo,
      tags,
      inspectionData
    } as PhotoMetadata;
  }
}
```

### Operational Transform for Collaborative Editing

**For Real-time Collaborative Features**:
```typescript
interface Operation {
  type: 'insert' | 'delete' | 'retain';
  chars?: string;
  count: number;
  attributes?: Record<string, unknown>;
}

export class OperationalTransform {
  static transform(op1: Operation[], op2: Operation[], priority: 'left' | 'right'): [Operation[], Operation[]] {
    const result1: Operation[] = [];
    const result2: Operation[] = [];
    
    let i = 0, j = 0;
    let offset1 = 0, offset2 = 0;

    while (i < op1.length && j < op2.length) {
      const o1 = op1[i];
      const o2 = op2[j];

      if (o1.type === 'retain' && o2.type === 'retain') {
        const minCount = Math.min(o1.count, o2.count);
        result1.push({ type: 'retain', count: minCount });
        result2.push({ type: 'retain', count: minCount });
        
        this.updateOperationCount(o1, minCount);
        this.updateOperationCount(o2, minCount);
        
        if (o1.count === 0) i++;
        if (o2.count === 0) j++;
        
      } else if (o1.type === 'insert') {
        result1.push(o1);
        result2.push({ type: 'retain', count: o1.chars?.length || 0 });
        i++;
        
      } else if (o2.type === 'insert') {
        if (priority === 'left') {
          result1.push({ type: 'retain', count: o2.chars?.length || 0 });
          result2.push(o2);
        } else {
          result1.push({ type: 'retain', count: o2.chars?.length || 0 });
          result2.push(o2);
        }
        j++;
        
      } else {
        // Handle delete operations
        const minCount = Math.min(o1.count, o2.count);
        
        this.updateOperationCount(o1, minCount);
        this.updateOperationCount(o2, minCount);
        
        if (o1.count === 0) i++;
        if (o2.count === 0) j++;
      }
    }

    // Add remaining operations
    result1.push(...op1.slice(i));
    result2.push(...op2.slice(j));

    return [result1, result2];
  }

  private static updateOperationCount(op: Operation, count: number): void {
    op.count -= count;
  }
}
```

---

## Network State Management

### Intelligent Sync Scheduling

```typescript
interface NetworkInfo {
  isConnected: boolean;
  type: 'wifi' | 'cellular' | 'none';
  isInternetReachable: boolean;
  details: {
    strength: number;
    frequency?: number;
    subnet?: string;
  };
}

interface SyncStrategy {
  priority: number;
  batchSize: number;
  maxRetries: number;
  backoffMultiplier: number;
  requiresWifi: boolean;
  requiresCharging?: boolean;
}

export class IntelligentSyncManager {
  private syncStrategies = new Map<string, SyncStrategy>();
  private currentNetwork: NetworkInfo | null = null;
  private syncQueue: SyncTask[] = [];
  private activeSyncs = new Map<string, Promise<void>>();

  constructor(private networkMonitor: NetworkMonitor) {
    this.setupNetworkMonitoring();
    this.configureSyncStrategies();
  }

  private configureSyncStrategies(): void {
    // High priority: Small metadata updates
    this.syncStrategies.set('metadata_update', {
      priority: 1,
      batchSize: 50,
      maxRetries: 5,
      backoffMultiplier: 2,
      requiresWifi: false
    });

    // Medium priority: Photo uploads
    this.syncStrategies.set('photo_upload', {
      priority: 2,
      batchSize: 3,
      maxRetries: 3,
      backoffMultiplier: 2,
      requiresWifi: true, // Large files prefer WiFi
      requiresCharging: false
    });

    // Low priority: Analytics and logs
    this.syncStrategies.set('analytics', {
      priority: 3,
      batchSize: 100,
      maxRetries: 2,
      backoffMultiplier: 1.5,
      requiresWifi: false
    });
  }

  private setupNetworkMonitoring(): void {
    this.networkMonitor.onNetworkChange((networkInfo) => {
      const wasOffline = !this.currentNetwork?.isConnected;
      this.currentNetwork = networkInfo;

      if (wasOffline && networkInfo.isConnected) {
        // Just came online - start syncing
        this.triggerSync();
      } else if (!networkInfo.isConnected) {
        // Went offline - pause active syncs
        this.pauseActiveSyncs();
      } else if (networkInfo.type === 'wifi' && this.currentNetwork?.type !== 'wifi') {
        // Switched to WiFi - resume high-bandwidth syncs
        this.resumeHighBandwidthSyncs();
      }
    });
  }

  async queueSync(task: Omit<SyncTask, 'id' | 'attempts' | 'nextAttempt'>): Promise<void> {
    const syncTask: SyncTask = {
      ...task,
      id: this.generateTaskId(),
      attempts: 0,
      nextAttempt: Date.now()
    };

    this.syncQueue.push(syncTask);
    this.sortSyncQueue();

    if (this.shouldSyncNow()) {
      this.triggerSync();
    }
  }

  private shouldSyncNow(): boolean {
    if (!this.currentNetwork?.isConnected) return false;

    const pendingTasks = this.syncQueue.filter(task => task.nextAttempt <= Date.now());
    
    return pendingTasks.some(task => {
      const strategy = this.syncStrategies.get(task.type);
      if (!strategy) return false;

      // Check network requirements
      if (strategy.requiresWifi && this.currentNetwork?.type !== 'wifi') {
        return false;
      }

      // Check charging requirement
      if (strategy.requiresCharging && !this.isBatteryCharging()) {
        return false;
      }

      return true;
    });
  }

  private async triggerSync(): Promise<void> {
    if (this.activeSyncs.size >= 3) return; // Limit concurrent syncs

    const readyTasks = this.getReadyTasks();
    
    for (const task of readyTasks) {
      if (this.activeSyncs.has(task.id)) continue;

      const syncPromise = this.executeSync(task)
        .finally(() => this.activeSyncs.delete(task.id));
      
      this.activeSyncs.set(task.id, syncPromise);
    }
  }

  private getReadyTasks(): SyncTask[] {
    return this.syncQueue
      .filter(task => {
        const strategy = this.syncStrategies.get(task.type);
        return strategy && 
               task.nextAttempt <= Date.now() && 
               this.meetsNetworkRequirements(strategy);
      })
      .slice(0, 5); // Limit batch size
  }

  private meetsNetworkRequirements(strategy: SyncStrategy): boolean {
    if (!this.currentNetwork?.isConnected) return false;

    if (strategy.requiresWifi && this.currentNetwork.type !== 'wifi') {
      return false;
    }

    if (strategy.requiresCharging && !this.isBatteryCharging()) {
      return false;
    }

    return true;
  }

  private async executeSync(task: SyncTask): Promise<void> {
    const strategy = this.syncStrategies.get(task.type);
    if (!strategy) throw new Error(`No strategy for task type: ${task.type}`);

    try {
      task.attempts++;
      
      switch (task.type) {
        case 'photo_upload':
          await this.syncPhotos(task.data, strategy);
          break;
        case 'metadata_update':
          await this.syncMetadata(task.data, strategy);
          break;
        case 'analytics':
          await this.syncAnalytics(task.data, strategy);
          break;
        default:
          throw new Error(`Unknown task type: ${task.type}`);
      }

      // Remove successful task
      this.syncQueue = this.syncQueue.filter(t => t.id !== task.id);

    } catch (error) {
      if (task.attempts >= strategy.maxRetries) {
        // Move to failed queue
        console.error(`Task ${task.id} failed after ${strategy.maxRetries} attempts:`, error);
        this.syncQueue = this.syncQueue.filter(t => t.id !== task.id);
      } else {
        // Schedule retry with exponential backoff
        const delay = 1000 * Math.pow(strategy.backoffMultiplier, task.attempts);
        task.nextAttempt = Date.now() + delay;
      }
    }
  }

  private sortSyncQueue(): void {
    this.syncQueue.sort((a, b) => {
      const strategyA = this.syncStrategies.get(a.type);
      const strategyB = this.syncStrategies.get(b.type);
      
      if (!strategyA || !strategyB) return 0;
      
      // Sort by priority first, then by next attempt time
      if (strategyA.priority !== strategyB.priority) {
        return strategyA.priority - strategyB.priority;
      }
      
      return a.nextAttempt - b.nextAttempt;
    });
  }
}
```

### Bandwidth-Aware Upload

```typescript
interface BandwidthMetrics {
  downloadSpeed: number; // Mbps
  uploadSpeed: number;   // Mbps
  latency: number;       // ms
  reliability: number;   // 0-1 (packet loss rate)
}

export class BandwidthAwareUploader {
  private bandwidthHistory: BandwidthMetrics[] = [];
  private currentStrategy: UploadStrategy = 'conservative';

  async uploadPhoto(photoData: PhotoFile, metadata: PhotoMetadata): Promise<UploadResult> {
    const metrics = await this.measureBandwidth();
    const strategy = this.selectUploadStrategy(metrics);

    switch (strategy) {
      case 'high_quality':
        return await this.uploadHighQuality(photoData, metadata);
      case 'balanced':
        return await this.uploadBalanced(photoData, metadata);
      case 'conservative':
        return await this.uploadConservative(photoData, metadata);
      case 'background_only':
        return await this.queueForBackgroundUpload(photoData, metadata);
    }
  }

  private async measureBandwidth(): Promise<BandwidthMetrics> {
    const startTime = Date.now();
    
    try {
      // Upload a small test file to measure speed
      const testData = new Uint8Array(1024 * 100); // 100KB test
      const uploadStart = Date.now();
      
      await this.uploadTestData(testData);
      
      const uploadTime = Date.now() - uploadStart;
      const uploadSpeed = (testData.length * 8) / (uploadTime * 1000); // Mbps
      
      const latency = Date.now() - startTime - uploadTime;

      const metrics: BandwidthMetrics = {
        downloadSpeed: 0, // Not measured in this example
        uploadSpeed,
        latency,
        reliability: this.calculateReliability()
      };

      this.bandwidthHistory.push(metrics);
      
      // Keep only recent measurements
      if (this.bandwidthHistory.length > 10) {
        this.bandwidthHistory.shift();
      }

      return metrics;
    } catch (error) {
      return {
        downloadSpeed: 0,
        uploadSpeed: 0.1, // Assume very slow connection
        latency: 5000,
        reliability: 0.1
      };
    }
  }

  private selectUploadStrategy(metrics: BandwidthMetrics): UploadStrategy {
    if (metrics.uploadSpeed > 10 && metrics.reliability > 0.9) {
      return 'high_quality';
    } else if (metrics.uploadSpeed > 2 && metrics.reliability > 0.7) {
      return 'balanced';
    } else if (metrics.uploadSpeed > 0.5 && metrics.reliability > 0.5) {
      return 'conservative';
    } else {
      return 'background_only';
    }
  }

  private async uploadHighQuality(photo: PhotoFile, metadata: PhotoMetadata): Promise<UploadResult> {
    // Upload full resolution immediately
    return await this.uploadWithChunking(photo, {
      quality: 100,
      maxChunkSize: 1024 * 1024, // 1MB chunks
      compression: false
    });
  }

  private async uploadBalanced(photo: PhotoFile, metadata: PhotoMetadata): Promise<UploadResult> {
    // Light compression, medium chunks
    const compressedPhoto = await this.compressPhoto(photo, 90);
    
    return await this.uploadWithChunking(compressedPhoto, {
      quality: 90,
      maxChunkSize: 512 * 1024, // 512KB chunks
      compression: true
    });
  }

  private async uploadConservative(photo: PhotoFile, metadata: PhotoMetadata): Promise<UploadResult> {
    // Heavy compression, small chunks
    const compressedPhoto = await this.compressPhoto(photo, 70);
    
    return await this.uploadWithChunking(compressedPhoto, {
      quality: 70,
      maxChunkSize: 256 * 1024, // 256KB chunks
      compression: true,
      retries: 5
    });
  }

  private async queueForBackgroundUpload(photo: PhotoFile, metadata: PhotoMetadata): Promise<UploadResult> {
    // Queue for later when network improves
    await this.backgroundUploadQueue.add({
      photo,
      metadata,
      priority: 'low',
      requiresWifi: true
    });

    return {
      success: true,
      queued: true,
      message: 'Queued for background upload'
    };
  }

  private async uploadWithChunking(
    photo: PhotoFile,
    options: ChunkUploadOptions
  ): Promise<UploadResult> {
    const fileSize = photo.size || 0;
    const chunkCount = Math.ceil(fileSize / options.maxChunkSize);
    
    const uploadSession = await this.initializeUploadSession({
      filename: photo.fileName || 'photo.jpg',
      fileSize,
      chunkCount,
      checksum: await this.calculateChecksum(photo)
    });

    try {
      for (let chunkIndex = 0; chunkIndex < chunkCount; chunkIndex++) {
        const start = chunkIndex * options.maxChunkSize;
        const end = Math.min(start + options.maxChunkSize, fileSize);
        
        const chunkData = await this.extractChunk(photo, start, end);
        
        await this.uploadChunk(uploadSession.sessionId, chunkIndex, chunkData, {
          retries: options.retries || 3
        });
      }

      const result = await this.finalizeUpload(uploadSession.sessionId);
      
      return {
        success: true,
        url: result.url,
        uploadedSize: fileSize,
        compressionRatio: options.compression ? 0.7 : 1.0
      };
    } catch (error) {
      await this.cancelUpload(uploadSession.sessionId);
      throw error;
    }
  }
}

type UploadStrategy = 'high_quality' | 'balanced' | 'conservative' | 'background_only';

interface ChunkUploadOptions {
  quality: number;
  maxChunkSize: number;
  compression: boolean;
  retries?: number;
}
```

---

## Conflict Resolution Strategies

### Last-Write-Wins (LWW)

```typescript
interface TimestampedRecord {
  id: string;
  data: Record<string, unknown>;
  timestamp: number;
  deviceId: string;
}

export class LastWriteWinsResolver {
  resolveConflict(local: TimestampedRecord, remote: TimestampedRecord): TimestampedRecord {
    // Simple timestamp comparison
    if (remote.timestamp > local.timestamp) {
      return remote;
    } else if (local.timestamp > remote.timestamp) {
      return local;
    } else {
      // Same timestamp - use device ID as tiebreaker
      return local.deviceId.localeCompare(remote.deviceId) < 0 ? local : remote;
    }
  }
}
```

### Multi-Value Register (MVR)

```typescript
interface VersionVector {
  [deviceId: string]: number;
}

interface VersionedValue<T> {
  value: T;
  vector: VersionVector;
  timestamp: number;
}

export class MultiValueRegister<T> {
  private values = new Set<VersionedValue<T>>();

  set(value: T, deviceId: string, timestamp?: number): void {
    const ts = timestamp || Date.now();
    
    // Remove obsolete values
    this.removeObsoleteValues(deviceId, ts);
    
    // Add new value
    const vector: VersionVector = { [deviceId]: ts };
    this.values.add({ value, vector, timestamp: ts });
  }

  merge(other: MultiValueRegister<T>): void {
    for (const otherValue of other.values) {
      let isDominated = false;
      const toRemove: VersionedValue<T>[] = [];

      for (const myValue of this.values) {
        const comparison = this.compareVectors(myValue.vector, otherValue.vector);
        
        if (comparison === 'dominates') {
          // My value is newer - ignore other value
          isDominated = true;
          break;
        } else if (comparison === 'dominated_by') {
          // Other value is newer - remove my value
          toRemove.push(myValue);
        }
        // If concurrent, keep both values
      }

      // Remove dominated values
      toRemove.forEach(value => this.values.delete(value));

      // Add other value if not dominated
      if (!isDominated) {
        this.values.add(otherValue);
      }
    }
  }

  getValues(): T[] {
    return Array.from(this.values).map(v => v.value);
  }

  hasConflict(): boolean {
    return this.values.size > 1;
  }

  private compareVectors(v1: VersionVector, v2: VersionVector): 'dominates' | 'dominated_by' | 'concurrent' {
    let v1Dominates = false;
    let v2Dominates = false;

    const allDevices = new Set([...Object.keys(v1), ...Object.keys(v2)]);

    for (const device of allDevices) {
      const t1 = v1[device] || 0;
      const t2 = v2[device] || 0;

      if (t1 > t2) v1Dominates = true;
      if (t2 > t1) v2Dominates = true;
    }

    if (v1Dominates && !v2Dominates) return 'dominates';
    if (v2Dominates && !v1Dominates) return 'dominated_by';
    return 'concurrent';
  }

  private removeObsoleteValues(deviceId: string, timestamp: number): void {
    this.values.forEach(value => {
      if (value.vector[deviceId] && value.vector[deviceId] < timestamp) {
        this.values.delete(value);
      }
    });
  }
}
```

### Domain-Specific Conflict Resolution

```typescript
interface PhotoConflict {
  photoId: string;
  localVersion: Photo;
  remoteVersion: Photo;
  conflictFields: string[];
}

export class ManufacturingConflictResolver {
  resolvePhotoConflict(conflict: PhotoConflict): Photo {
    const resolution = { ...conflict.localVersion };

    conflict.conflictFields.forEach(field => {
      switch (field) {
        case 'metadata.qualityNotes':
          // For quality notes, merge both versions
          resolution.metadata.qualityNotes = this.mergeQualityNotes(
            conflict.localVersion.metadata.qualityNotes,
            conflict.remoteVersion.metadata.qualityNotes
          );
          break;

        case 'metadata.tags':
          // Tags are additive - merge both sets
          resolution.metadata.tags = this.mergeTags(
            conflict.localVersion.metadata.tags,
            conflict.remoteVersion.metadata.tags
          );
          break;

        case 'syncStatus':
          // Always prefer 'synced' status
          resolution.syncStatus = this.resolveStatus(
            conflict.localVersion.syncStatus,
            conflict.remoteVersion.syncStatus
          );
          break;

        case 'metadata.inspectionResults':
          // For inspection results, use the most recent
          resolution.metadata.inspectionResults = this.resolveInspectionResults(
            conflict.localVersion,
            conflict.remoteVersion
          );
          break;

        default:
          // For other fields, use Last-Write-Wins
          if (conflict.remoteVersion.lastModified > conflict.localVersion.lastModified) {
            (resolution as any)[field] = (conflict.remoteVersion as any)[field];
          }
      }
    });

    // Update conflict resolution metadata
    resolution.conflictVersion = (resolution.conflictVersion || 0) + 1;
    resolution.lastModified = Date.now();

    return resolution;
  }

  private mergeQualityNotes(local?: string, remote?: string): string {
    if (!local && !remote) return '';
    if (!local) return remote!;
    if (!remote) return local;

    // Merge notes with attribution
    return `${local}\n\n--- Remote changes ---\n${remote}`;
  }

  private mergeTags(localTags: string[] = [], remoteTags: string[] = []): string[] {
    const tagSet = new Set([...localTags, ...remoteTags]);
    return Array.from(tagSet).sort();
  }

  private resolveStatus(localStatus: SyncStatus, remoteStatus: SyncStatus): SyncStatus {
    const statusPriority: Record<SyncStatus, number> = {
      'synced': 4,
      'syncing': 3,
      'pending': 2,
      'failed': 1,
      'conflicted': 0
    };

    return statusPriority[localStatus] > statusPriority[remoteStatus] 
      ? localStatus 
      : remoteStatus;
  }

  private resolveInspectionResults(local: Photo, remote: Photo): any {
    // Use the inspection with the most recent timestamp
    const localInspection = local.metadata.inspectionResults;
    const remoteInspection = remote.metadata.inspectionResults;

    if (!localInspection) return remoteInspection;
    if (!remoteInspection) return localInspection;

    return localInspection.timestamp > remoteInspection.timestamp
      ? localInspection
      : remoteInspection;
  }
}
```

---

## Performance Optimization for Large Datasets

### Pagination and Virtual Scrolling

```typescript
interface PaginatedQuery<T> {
  items: T[];
  hasMore: boolean;
  nextCursor?: string;
  totalCount?: number;
}

export class OfflineFirstPagination<T> {
  private cache = new Map<string, T[]>();
  private pageSize = 50;

  async getPaginatedResults(
    query: string,
    cursor?: string,
    useCache = true
  ): Promise<PaginatedQuery<T>> {
    const cacheKey = `${query}_${cursor || 'first'}`;
    
    // Check cache first
    if (useCache && this.cache.has(cacheKey)) {
      const cachedItems = this.cache.get(cacheKey)!;
      return {
        items: cachedItems,
        hasMore: cachedItems.length === this.pageSize,
        nextCursor: this.generateNextCursor(cachedItems)
      };
    }

    // Query local database
    const localResults = await this.queryLocalDatabase(query, cursor);
    
    // Cache results
    this.cache.set(cacheKey, localResults.items);
    
    // Trigger background sync if needed
    if (this.shouldTriggerBackgroundSync(localResults)) {
      this.triggerBackgroundSync(query, cursor);
    }

    return localResults;
  }

  private async queryLocalDatabase(
    query: string,
    cursor?: string
  ): Promise<PaginatedQuery<T>> {
    // Implementation depends on database choice
    // Example with WatermelonDB:
    const baseQuery = this.buildWatermelonQuery(query);
    
    if (cursor) {
      // Add cursor-based filtering
      baseQuery.append(Q.where('id', Q.gt(cursor)));
    }

    const items = await baseQuery
      .append(Q.sortBy('created_at', Q.desc))
      .append(Q.take(this.pageSize + 1)) // Get one extra to check for more
      .fetch();

    const hasMore = items.length > this.pageSize;
    const resultItems = hasMore ? items.slice(0, -1) : items;

    return {
      items: resultItems as T[],
      hasMore,
      nextCursor: hasMore ? resultItems[resultItems.length - 1].id : undefined
    };
  }

  private shouldTriggerBackgroundSync(results: PaginatedQuery<T>): boolean {
    // Trigger sync if we have few results or they're old
    return results.items.length < this.pageSize / 2 ||
           this.areResultsStale(results.items);
  }

  private areResultsStale(items: T[]): boolean {
    if (items.length === 0) return true;
    
    const newest = items[0] as any;
    const age = Date.now() - (newest.updated_at || newest.created_at);
    
    return age > 300000; // 5 minutes
  }
}
```

### Incremental Sync

```typescript
interface SyncCheckpoint {
  tableName: string;
  lastSyncTimestamp: number;
  syncVersion: number;
  checksum: string;
}

export class IncrementalSyncManager {
  private checkpoints = new Map<string, SyncCheckpoint>();

  async performIncrementalSync(tableName: string): Promise<SyncResult> {
    const checkpoint = this.checkpoints.get(tableName);
    const lastSync = checkpoint?.lastSyncTimestamp || 0;

    // Get local changes since last sync
    const localChanges = await this.getLocalChangesSince(tableName, lastSync);
    
    // Get remote changes since last sync
    const remoteChanges = await this.getRemoteChangesSince(tableName, lastSync);

    // Apply remote changes locally
    const applyResults = await this.applyRemoteChanges(tableName, remoteChanges);

    // Upload local changes
    const uploadResults = await this.uploadLocalChanges(tableName, localChanges);

    // Update checkpoint
    const newCheckpoint: SyncCheckpoint = {
      tableName,
      lastSyncTimestamp: Date.now(),
      syncVersion: (checkpoint?.syncVersion || 0) + 1,
      checksum: await this.calculateTableChecksum(tableName)
    };
    
    this.checkpoints.set(tableName, newCheckpoint);
    await this.persistCheckpoint(newCheckpoint);

    return {
      table: tableName,
      localChanges: localChanges.length,
      remoteChanges: remoteChanges.length,
      conflicts: applyResults.conflicts,
      uploadErrors: uploadResults.errors,
      checkpoint: newCheckpoint
    };
  }

  private async getLocalChangesSince(
    tableName: string, 
    timestamp: number
  ): Promise<ChangeRecord[]> {
    // Query local change log or use last_modified timestamps
    const changes = await this.database.execute(`
      SELECT * FROM ${tableName} 
      WHERE last_modified > ? AND sync_status IN ('pending', 'failed')
      ORDER BY last_modified ASC
    `, [timestamp]);

    return changes.rows.map(row => ({
      id: row.id,
      operation: this.determineOperation(row),
      data: row,
      timestamp: row.last_modified
    }));
  }

  private async getRemoteChangesSince(
    tableName: string,
    timestamp: number
  ): Promise<ChangeRecord[]> {
    const response = await this.apiClient.get(`/sync/${tableName}/changes`, {
      params: { since: timestamp, limit: 1000 }
    });

    return response.data.changes;
  }

  private async applyRemoteChanges(
    tableName: string,
    changes: ChangeRecord[]
  ): Promise<ApplyResult> {
    const conflicts: ConflictRecord[] = [];
    let applied = 0;

    for (const change of changes) {
      try {
        const localRecord = await this.database.get(tableName).find(change.id);
        
        if (localRecord && localRecord.lastModified > change.timestamp) {
          // Local version is newer - conflict
          conflicts.push({
            id: change.id,
            localVersion: localRecord,
            remoteVersion: change.data,
            field: this.findConflictingFields(localRecord, change.data)
          });
        } else {
          // Apply remote change
          await this.applyChange(tableName, change);
          applied++;
        }
      } catch (error) {
        console.error(`Failed to apply change ${change.id}:`, error);
      }
    }

    return { applied, conflicts };
  }

  private async calculateTableChecksum(tableName: string): Promise<string> {
    // Calculate checksum of all records for integrity verification
    const result = await this.database.execute(`
      SELECT GROUP_CONCAT(id || '|' || last_modified) as concat_data
      FROM ${tableName}
      ORDER BY id
    `);

    const data = result.rows[0]?.concat_data || '';
    
    // Simple hash - in production use crypto hash
    return Buffer.from(data).toString('base64').substr(0, 16);
  }
}
```

---

## Testing Offline-First Applications

### Network Simulation

```typescript
export class NetworkSimulator {
  private isOnline = true;
  private latency = 0;
  private reliability = 1.0;
  private bandwidth = Infinity;

  simulate(scenario: NetworkScenario): void {
    switch (scenario) {
      case 'offline':
        this.isOnline = false;
        break;
      case 'slow_3g':
        this.isOnline = true;
        this.latency = 400;
        this.bandwidth = 0.4; // Mbps
        this.reliability = 0.9;
        break;
      case 'fast_3g':
        this.isOnline = true;
        this.latency = 200;
        this.bandwidth = 1.6;
        this.reliability = 0.95;
        break;
      case 'wifi':
        this.isOnline = true;
        this.latency = 50;
        this.bandwidth = 50;
        this.reliability = 0.99;
        break;
      case 'intermittent':
        this.simulateIntermittentConnection();
        break;
    }
  }

  async simulateRequest<T>(request: () => Promise<T>): Promise<T> {
    if (!this.isOnline) {
      throw new Error('Network unavailable');
    }

    // Simulate latency
    if (this.latency > 0) {
      await this.delay(this.latency + Math.random() * this.latency * 0.5);
    }

    // Simulate reliability
    if (Math.random() > this.reliability) {
      throw new Error('Network request failed');
    }

    // Simulate bandwidth limitations
    const startTime = Date.now();
    const result = await request();
    const minDuration = this.calculateMinDuration(result);
    
    const elapsed = Date.now() - startTime;
    if (elapsed < minDuration) {
      await this.delay(minDuration - elapsed);
    }

    return result;
  }

  private simulateIntermittentConnection(): void {
    const toggle = () => {
      this.isOnline = !this.isOnline;
      const nextToggle = Math.random() * 10000 + 5000; // 5-15 seconds
      setTimeout(toggle, nextToggle);
    };
    
    toggle();
  }

  private calculateMinDuration(data: unknown): number {
    const dataSize = JSON.stringify(data).length;
    const bytesPerSecond = (this.bandwidth * 1024 * 1024) / 8;
    return (dataSize / bytesPerSecond) * 1000;
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

type NetworkScenario = 'offline' | 'slow_3g' | 'fast_3g' | 'wifi' | 'intermittent';
```

### Integration Tests

```typescript
describe('Offline-First Sync Integration', () => {
  let syncManager: IncrementalSyncManager;
  let networkSimulator: NetworkSimulator;
  let database: Database;

  beforeEach(async () => {
    database = await createTestDatabase();
    syncManager = new IncrementalSyncManager(database);
    networkSimulator = new NetworkSimulator();
  });

  test('should handle offline photo capture and eventual sync', async () => {
    // Simulate offline state
    networkSimulator.simulate('offline');
    
    // Capture photos while offline
    const photos = [];
    for (let i = 0; i < 5; i++) {
      const photo = await captureTestPhoto({
        department: 'Manufacturing',
        project: `Project${i}`
      });
      photos.push(photo);
    }

    // Verify photos stored locally
    const localPhotos = await database.get('photos').query().fetch();
    expect(localPhotos).toHaveLength(5);
    expect(localPhotos.every(p => p.syncStatus === 'pending')).toBe(true);

    // Come back online
    networkSimulator.simulate('wifi');

    // Trigger sync
    const syncResult = await syncManager.performIncrementalSync('photos');
    
    // Verify sync results
    expect(syncResult.localChanges).toBe(5);
    expect(syncResult.uploadErrors).toHaveLength(0);

    // Verify photos marked as synced
    const syncedPhotos = await database.get('photos').query().fetch();
    expect(syncedPhotos.every(p => p.syncStatus === 'synced')).toBe(true);
  });

  test('should resolve conflicts correctly', async () => {
    // Create a photo
    const photo = await captureTestPhoto({ department: 'QC' });
    
    // Sync initially
    networkSimulator.simulate('wifi');
    await syncManager.performIncrementalSync('photos');

    // Simulate concurrent modifications
    networkSimulator.simulate('offline');
    
    // Local modification
    await photo.update(p => {
      p.metadata = { ...p.metadata, localNote: 'Local change' };
      p.lastModified = Date.now();
    });

    // Simulate remote modification (would happen on another device)
    const remoteChange = {
      id: photo.id,
      timestamp: Date.now() + 1000, // Newer timestamp
      data: {
        ...photo._raw,
        metadata: JSON.stringify({ 
          ...photo.metadata, 
          remoteNote: 'Remote change' 
        })
      }
    };

    // Come online and sync
    networkSimulator.simulate('wifi');
    
    // Mock remote changes
    jest.spyOn(syncManager as any, 'getRemoteChangesSince')
      .mockResolvedValue([remoteChange]);

    const syncResult = await syncManager.performIncrementalSync('photos');
    
    // Verify conflict resolution
    expect(syncResult.conflicts).toHaveLength(1);
    
    // Verify conflict was resolved (remote wins due to newer timestamp)
    const resolvedPhoto = await database.get('photos').find(photo.id);
    expect(resolvedPhoto.metadata.remoteNote).toBe('Remote change');
  });

  test('should handle intermittent connectivity gracefully', async () => {
    // Start with intermittent connection
    networkSimulator.simulate('intermittent');
    
    // Attempt to sync multiple times
    const syncPromises = Array.from({ length: 10 }, () => 
      syncManager.performIncrementalSync('photos')
        .catch(error => ({ error: error.message }))
    );

    const results = await Promise.allSettled(syncPromises);
    
    // Some should succeed, some should fail
    const successful = results.filter(r => 
      r.status === 'fulfilled' && !(r.value as any).error
    );
    const failed = results.filter(r => 
      r.status === 'rejected' || (r.value as any).error
    );

    expect(successful.length).toBeGreaterThan(0);
    expect(failed.length).toBeGreaterThan(0);
    
    // Eventually all should sync when connection stabilizes
    networkSimulator.simulate('wifi');
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const finalSync = await syncManager.performIncrementalSync('photos');
    expect(finalSync.uploadErrors).toHaveLength(0);
  });
});
```

This comprehensive guide provides production-ready patterns for building robust offline-first applications with reliable synchronization, conflict resolution, and performance optimization for enterprise use cases.
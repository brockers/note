# MongoDB TypeScript Integration Patterns for Mobile Applications (2024-2025)

**Comprehensive guide for integrating MongoDB with TypeScript mobile applications, focusing on offline-first architectures, real-time synchronization, and enterprise-scale data management.**

---

## MongoDB Integration Approaches

### Direct MongoDB Connection (Server-Side Only)

**Use Case**: Backend services, API servers, Node.js applications
- **Security**: Database credentials never exposed to clients
- **Performance**: Direct connection with connection pooling
- **Flexibility**: Full MongoDB feature access

**TypeScript Implementation**:
```typescript
import { MongoClient, Db, Collection, MongoClientOptions } from 'mongodb';
import { z } from 'zod';

// Configuration with type safety
const MongoConfigSchema = z.object({
  connectionString: z.string().url(),
  databaseName: z.string().min(1),
  maxPoolSize: z.number().default(10),
  serverSelectionTimeoutMS: z.number().default(5000),
  connectTimeoutMS: z.number().default(10000)
});

type MongoConfig = z.infer<typeof MongoConfigSchema>;

export class MongoDBService {
  private client: MongoClient;
  private db: Db;
  private isConnected = false;

  constructor(private config: MongoConfig) {
    const options: MongoClientOptions = {
      maxPoolSize: config.maxPoolSize,
      serverSelectionTimeoutMS: config.serverSelectionTimeoutMS,
      connectTimeoutMS: config.connectTimeoutMS,
      retryWrites: true,
      writeConcern: { w: 'majority', j: true },
      readPreference: 'primary'
    };

    this.client = new MongoClient(config.connectionString, options);
  }

  async connect(): Promise<void> {
    if (this.isConnected) return;

    try {
      await this.client.connect();
      this.db = this.client.db(this.config.databaseName);
      this.isConnected = true;
      
      console.log(`Connected to MongoDB database: ${this.config.databaseName}`);
    } catch (error) {
      console.error('Failed to connect to MongoDB:', error);
      throw error;
    }
  }

  async disconnect(): Promise<void> {
    if (!this.isConnected) return;

    await this.client.close();
    this.isConnected = false;
  }

  getCollection<T = Document>(name: string): Collection<T> {
    if (!this.isConnected) {
      throw new Error('Database not connected');
    }
    return this.db.collection<T>(name);
  }

  async healthCheck(): Promise<{ status: string; latency: number }> {
    const start = Date.now();
    
    try {
      await this.db.admin().ping();
      return {
        status: 'healthy',
        latency: Date.now() - start
      };
    } catch (error) {
      return {
        status: 'unhealthy',
        latency: Date.now() - start
      };
    }
  }
}
```

### MongoDB Atlas Device Sync (Deprecated - Migration Required)

**Status**: Atlas Device Sync is being deprecated in 2024
- **Migration Path**: Move to PowerSync, custom sync solutions, or GraphQL APIs
- **Legacy Support**: Existing applications need migration planning

**Legacy Pattern** (for reference only):
```typescript
// ⚠️ DEPRECATED - Atlas Device Sync pattern
import Realm from 'realm';

// This pattern is no longer recommended for new projects
const PhotoSchema = {
  name: 'Photo',
  primaryKey: '_id',
  properties: {
    _id: 'objectId',
    localPath: 'string',
    remoteUrl: 'string?',
    departmentId: 'string',
    projectId: 'string',
    metadata: 'mixed',
    _partition: 'string' // Required for Device Sync
  }
};

// Migration to PowerSync or custom sync is recommended
```

### PowerSync + MongoDB (Recommended)

**Architecture**: SQLite local database with real-time MongoDB synchronization
- **Local Storage**: SQLite with reactive queries
- **Sync Engine**: PowerSync handles MongoDB change streams
- **Conflict Resolution**: Built-in with customizable strategies

**Setup and Configuration**:
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
      new Column({ name: 'part_id', type: ColumnType.text }),
      new Column({ name: 'captured_at', type: ColumnType.integer }),
      new Column({ name: 'sync_status', type: ColumnType.text }),
      new Column({ name: 'metadata_json', type: ColumnType.text }),
      new Column({ name: 'file_size', type: ColumnType.integer }),
      new Column({ name: 'checksum', type: ColumnType.text })
    ]
  }),
  new Table({
    name: 'departments',
    columns: [
      new Column({ name: 'name', type: ColumnType.text }),
      new Column({ name: 'code', type: ColumnType.text }),
      new Column({ name: 'active', type: ColumnType.integer }),
      new Column({ name: 'sync_policy', type: ColumnType.text })
    ]
  }),
  new Table({
    name: 'projects',
    columns: [
      new Column({ name: 'job_number', type: ColumnType.text }),
      new Column({ name: 'name', type: ColumnType.text }),
      new Column({ name: 'client', type: ColumnType.text }),
      new Column({ name: 'status', type: ColumnType.text }),
      new Column({ name: 'created_at', type: ColumnType.integer })
    ]
  })
]);

export class PowerSyncService {
  private db: PowerSyncDatabase;
  private uploadQueue = new Map<string, UploadTask>();
  private syncConfig: PowerSyncConfig;

  constructor(config: PowerSyncConfig) {
    this.syncConfig = config;
    this.db = new PowerSyncDatabase({
      schema,
      database: {
        dbFilename: 'manufacturing.sqlite',
        // Enable WAL mode for better concurrency
        flags: {
          enableWalMode: true,
          enableForeignKeys: true
        }
      }
    });
  }

  async initialize(): Promise<void> {
    await this.db.init();
    
    // Connect to PowerSync service
    await this.db.connect({
      powerSyncUrl: this.syncConfig.powerSyncUrl,
      token: await this.getAuthToken(),
      params: {
        user_id: await this.getCurrentUserId(),
        department_id: await this.getCurrentDepartmentId()
      }
    });

    this.setupUploadQueue();
    this.startSync();
  }

  private async getAuthToken(): Promise<string> {
    // Implement your authentication logic
    const response = await fetch(`${this.syncConfig.authEndpoint}/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        device_id: await this.getDeviceId(),
        department_id: await this.getCurrentDepartmentId()
      })
    });

    const { token } = await response.json();
    return token;
  }

  private setupUploadQueue(): void {
    // Handle large file uploads separately from PowerSync
    setInterval(async () => {
      await this.processUploadQueue();
    }, 10000); // Process every 10 seconds
  }

  private async processUploadQueue(): Promise<void> {
    const pendingUploads = Array.from(this.uploadQueue.values())
      .filter(task => task.status === 'pending')
      .slice(0, 3); // Process 3 uploads concurrently

    for (const task of pendingUploads) {
      this.uploadFile(task);
    }
  }

  async queuePhotoUpload(photoId: string, localPath: string): Promise<void> {
    const task: UploadTask = {
      id: photoId,
      type: 'photo',
      localPath,
      status: 'pending',
      attempts: 0,
      createdAt: Date.now()
    };

    this.uploadQueue.set(photoId, task);
  }

  private async uploadFile(task: UploadTask): Promise<void> {
    task.status = 'uploading';
    task.attempts++;

    try {
      const fileData = await this.readFile(task.localPath);
      const uploadResult = await this.uploadToStorage(fileData, task);

      // Update PowerSync database with remote URL
      await this.db.execute(`
        UPDATE photos 
        SET remote_url = ?, sync_status = 'synced' 
        WHERE id = ?
      `, [uploadResult.url, task.id]);

      this.uploadQueue.delete(task.id);
      task.status = 'completed';

    } catch (error) {
      task.status = 'failed';
      
      if (task.attempts >= 3) {
        console.error(`Upload failed permanently for ${task.id}:`, error);
        this.uploadQueue.delete(task.id);
      } else {
        // Retry with exponential backoff
        setTimeout(() => {
          task.status = 'pending';
        }, Math.pow(2, task.attempts) * 1000);
      }
    }
  }
}

interface PowerSyncConfig {
  powerSyncUrl: string;
  authEndpoint: string;
  uploadEndpoint: string;
}

interface UploadTask {
  id: string;
  type: 'photo' | 'document';
  localPath: string;
  status: 'pending' | 'uploading' | 'completed' | 'failed';
  attempts: number;
  createdAt: number;
}
```

---

## Type-Safe Data Models and Schemas

### Zod Schema Integration with MongoDB

**Benefits**:
- Runtime validation of all external data
- Type inference for TypeScript
- Automatic parsing and validation
- Integration with form libraries

```typescript
import { z } from 'zod';
import { ObjectId } from 'mongodb';

// Custom Zod schema for ObjectId
const ObjectIdSchema = z.custom<ObjectId>((val) => {
  return ObjectId.isValid(val);
}, 'Invalid ObjectId');

// Branded types for type safety
const DepartmentIdSchema = z.string().brand<'DepartmentId'>();
const ProjectIdSchema = ObjectIdSchema.brand<'ProjectId'>();
const PartIdSchema = z.string().brand<'PartId'>();
const PhotoIdSchema = ObjectIdSchema.brand<'PhotoId'>();

export type DepartmentId = z.infer<typeof DepartmentIdSchema>;
export type ProjectId = z.infer<typeof ProjectIdSchema>;
export type PartId = z.infer<typeof PartIdSchema>;
export type PhotoId = z.infer<typeof PhotoIdSchema>;

// Manufacturing photo schema with comprehensive validation
export const PhotoMetadataSchema = z.object({
  _id: PhotoIdSchema,
  localPath: z.string().min(1),
  remoteUrl: z.string().url().optional(),
  fileName: z.string().min(1),
  
  // Manufacturing context
  departmentId: DepartmentIdSchema,
  projectId: ProjectIdSchema,
  partId: PartIdSchema,
  jobNumber: z.string().min(1),
  
  // Timestamps
  capturedAt: z.date(),
  syncedAt: z.date().optional(),
  lastModified: z.date(),
  
  // File information
  fileSize: z.number().positive(),
  mimeType: z.string().regex(/^image\//),
  checksumMD5: z.string().length(32),
  
  // Sync status
  syncStatus: z.enum(['pending', 'syncing', 'synced', 'failed', 'conflicted']),
  syncAttempts: z.number().default(0),
  
  // Manufacturing-specific metadata
  metadata: z.object({
    inspectionType: z.string(),
    qualityRating: z.number().min(1).max(10).optional(),
    defectsFound: z.array(z.string()).default([]),
    inspectorNotes: z.string().optional(),
    machineId: z.string().optional(),
    batchNumber: z.string().optional(),
    
    // Location data
    location: z.object({
      latitude: z.number().min(-90).max(90),
      longitude: z.number().min(-180).max(180),
      accuracy: z.number().positive().optional()
    }).optional(),
    
    // Device information
    deviceInfo: z.object({
      deviceId: z.string(),
      model: z.string(),
      osVersion: z.string(),
      appVersion: z.string(),
      cameraSettings: z.object({
        quality: z.number().min(0).max(100),
        flash: z.enum(['auto', 'on', 'off']),
        resolution: z.object({
          width: z.number().positive(),
          height: z.number().positive()
        })
      })
    })
  }),
  
  // Audit trail
  auditTrail: z.array(z.object({
    action: z.enum(['created', 'updated', 'synced', 'deleted']),
    timestamp: z.date(),
    userId: z.string(),
    deviceId: z.string(),
    details: z.record(z.unknown()).optional()
  })).default([])
});

export type PhotoMetadata = z.infer<typeof PhotoMetadataSchema>;

// Department schema
export const DepartmentSchema = z.object({
  _id: ObjectIdSchema,
  name: z.enum(['Paint', 'Built Up/Structural', 'Panel/Trim', 'Z-Line', 'QA/QC']),
  code: z.string().length(3),
  active: z.boolean().default(true),
  
  // Sync configuration
  syncPolicy: z.enum(['realtime', 'batch', 'manual']).default('batch'),
  syncFrequencyMinutes: z.number().positive().default(15),
  
  // Department-specific settings
  settings: z.object({
    maxPhotosPerSession: z.number().positive().default(100),
    compressionQuality: z.number().min(50).max(100).default(90),
    requireGPSLocation: z.boolean().default(false),
    mandatoryFields: z.array(z.string()).default([]),
    photoRetentionDays: z.number().positive().default(365)
  }),
  
  createdAt: z.date(),
  updatedAt: z.date()
});

export type Department = z.infer<typeof DepartmentSchema>;

// Project schema
export const ProjectSchema = z.object({
  _id: ProjectIdSchema,
  jobNumber: z.string().min(1),
  name: z.string().min(1),
  client: z.string().min(1),
  
  status: z.enum(['active', 'completed', 'on-hold', 'cancelled']),
  priority: z.enum(['low', 'medium', 'high', 'critical']).default('medium'),
  
  // Associated departments
  departments: z.array(DepartmentIdSchema),
  
  // Project timeline
  startDate: z.date(),
  expectedEndDate: z.date(),
  actualEndDate: z.date().optional(),
  
  // Parts associated with this project
  parts: z.array(z.object({
    id: PartIdSchema,
    name: z.string(),
    partNumber: z.string(),
    description: z.string().optional(),
    requiredPhotos: z.number().default(1),
    specifications: z.record(z.unknown()).optional()
  })),
  
  // Project metadata
  metadata: z.object({
    estimatedPhotos: z.number().positive().optional(),
    actualPhotos: z.number().default(0),
    qualityScore: z.number().min(0).max(10).optional(),
    notes: z.string().optional()
  }),
  
  createdAt: z.date(),
  updatedAt: z.date()
});

export type Project = z.infer<typeof ProjectSchema>;

// Validation functions
export function validatePhotoMetadata(data: unknown): PhotoMetadata {
  return PhotoMetadataSchema.parse(data);
}

export function validateDepartment(data: unknown): Department {
  return DepartmentSchema.parse(data);
}

export function validateProject(data: unknown): Project {
  return ProjectSchema.parse(data);
}

// Partial validation for updates
export const PhotoMetadataUpdateSchema = PhotoMetadataSchema.partial().extend({
  _id: PhotoIdSchema, // ID is always required for updates
  lastModified: z.date().default(() => new Date())
});

export type PhotoMetadataUpdate = z.infer<typeof PhotoMetadataUpdateSchema>;
```

### Repository Pattern with MongoDB

```typescript
import { Collection, Filter, UpdateFilter, FindOptions } from 'mongodb';

export interface Repository<T, K = string> {
  create(data: Omit<T, '_id'>): Promise<T>;
  findById(id: K): Promise<T | null>;
  findMany(filter: Filter<T>, options?: FindOptions<T>): Promise<T[]>;
  update(id: K, updates: UpdateFilter<T>): Promise<T | null>;
  delete(id: K): Promise<boolean>;
  count(filter?: Filter<T>): Promise<number>;
}

export class MongoRepository<T extends { _id: ObjectId }, K = ObjectId> implements Repository<T, K> {
  constructor(
    private collection: Collection<T>,
    private validator: (data: unknown) => T,
    private partialValidator?: (data: unknown) => Partial<T>
  ) {}

  async create(data: Omit<T, '_id'>): Promise<T> {
    const docWithId = {
      ...data,
      _id: new ObjectId(),
      createdAt: new Date(),
      updatedAt: new Date()
    } as T;

    // Validate before insertion
    const validatedDoc = this.validator(docWithId);
    
    const result = await this.collection.insertOne(validatedDoc);
    
    if (!result.acknowledged) {
      throw new Error('Failed to create document');
    }

    return validatedDoc;
  }

  async findById(id: K): Promise<T | null> {
    const doc = await this.collection.findOne({ _id: id } as Filter<T>);
    
    if (!doc) return null;
    
    // Validate document from database
    return this.validator(doc);
  }

  async findMany(filter: Filter<T> = {}, options: FindOptions<T> = {}): Promise<T[]> {
    const cursor = this.collection.find(filter, options);
    const docs = await cursor.toArray();
    
    // Validate all documents
    return docs.map(doc => this.validator(doc));
  }

  async update(id: K, updates: UpdateFilter<T>): Promise<T | null> {
    const updateDoc = {
      ...updates,
      $set: {
        ...(updates.$set || {}),
        updatedAt: new Date()
      }
    };

    const result = await this.collection.findOneAndUpdate(
      { _id: id } as Filter<T>,
      updateDoc,
      { returnDocument: 'after' }
    );

    if (!result.value) return null;
    
    return this.validator(result.value);
  }

  async delete(id: K): Promise<boolean> {
    const result = await this.collection.deleteOne({ _id: id } as Filter<T>);
    return result.deletedCount === 1;
  }

  async count(filter: Filter<T> = {}): Promise<number> {
    return await this.collection.countDocuments(filter);
  }

  // Manufacturing-specific queries
  async findByDepartment(departmentId: DepartmentId): Promise<T[]> {
    return await this.findMany({ 
      departmentId 
    } as Filter<T>);
  }

  async findPendingSync(): Promise<T[]> {
    return await this.findMany({ 
      syncStatus: { $in: ['pending', 'failed'] }
    } as Filter<T>);
  }

  async findByDateRange(startDate: Date, endDate: Date): Promise<T[]> {
    return await this.findMany({
      capturedAt: {
        $gte: startDate,
        $lte: endDate
      }
    } as Filter<T>);
  }
}

// Specialized Photo Repository
export class PhotoRepository extends MongoRepository<PhotoMetadata, ObjectId> {
  constructor(collection: Collection<PhotoMetadata>) {
    super(collection, validatePhotoMetadata);
  }

  async findByProject(projectId: ProjectId): Promise<PhotoMetadata[]> {
    return await this.findMany({ projectId });
  }

  async findByPart(partId: PartId): Promise<PhotoMetadata[]> {
    return await this.findMany({ partId });
  }

  async getQualityStats(departmentId?: DepartmentId): Promise<QualityStats> {
    const matchStage: any = {};
    if (departmentId) {
      matchStage.departmentId = departmentId;
    }

    const pipeline = [
      { $match: matchStage },
      {
        $group: {
          _id: '$metadata.qualityRating',
          count: { $sum: 1 },
          avgFileSize: { $avg: '$fileSize' }
        }
      },
      { $sort: { _id: 1 } }
    ];

    const results = await this.collection.aggregate(pipeline).toArray();
    
    return {
      totalPhotos: results.reduce((sum, r) => sum + r.count, 0),
      qualityDistribution: results.map(r => ({
        rating: r._id,
        count: r.count,
        avgFileSize: r.avgFileSize
      })),
      averageQuality: this.calculateAverageQuality(results)
    };
  }

  private calculateAverageQuality(results: any[]): number {
    const totalWeighted = results.reduce((sum, r) => sum + (r._id * r.count), 0);
    const totalCount = results.reduce((sum, r) => sum + r.count, 0);
    return totalCount > 0 ? totalWeighted / totalCount : 0;
  }
}

interface QualityStats {
  totalPhotos: number;
  qualityDistribution: Array<{
    rating: number;
    count: number;
    avgFileSize: number;
  }>;
  averageQuality: number;
}
```

---

## Real-Time Synchronization with Change Streams

### MongoDB Change Streams

```typescript
import { ChangeStream, ChangeStreamDocument, ResumeToken } from 'mongodb';

interface ChangeStreamConfig {
  resumeAfter?: ResumeToken;
  startAtOperationTime?: Date;
  batchSize?: number;
  maxAwaitTimeMS?: number;
}

export class MongoChangeStreamManager {
  private changeStreams = new Map<string, ChangeStream>();
  private resumeTokens = new Map<string, ResumeToken>();
  private isListening = false;

  constructor(private mongoService: MongoDBService) {}

  async startListening(collections: string[], config: ChangeStreamConfig = {}): Promise<void> {
    if (this.isListening) return;

    for (const collectionName of collections) {
      await this.setupChangeStream(collectionName, config);
    }

    this.isListening = true;
  }

  private async setupChangeStream(
    collectionName: string, 
    config: ChangeStreamConfig
  ): Promise<void> {
    const collection = this.mongoService.getCollection(collectionName);
    
    const pipeline = [
      {
        $match: {
          'operationType': { $in: ['insert', 'update', 'delete'] },
          // Only listen to changes from specific departments
          'fullDocument.departmentId': { $exists: true }
        }
      }
    ];

    const options = {
      fullDocument: 'updateLookup' as const,
      batchSize: config.batchSize || 10,
      maxAwaitTimeMS: config.maxAwaitTimeMS || 10000,
      ...(config.resumeAfter && { resumeAfter: config.resumeAfter }),
      ...(config.startAtOperationTime && { startAtOperationTime: config.startAtOperationTime })
    };

    const changeStream = collection.watch(pipeline, options);
    this.changeStreams.set(collectionName, changeStream);

    changeStream.on('change', (change: ChangeStreamDocument) => {
      this.handleChange(collectionName, change);
    });

    changeStream.on('error', (error) => {
      console.error(`Change stream error for ${collectionName}:`, error);
      this.handleChangeStreamError(collectionName, error);
    });

    changeStream.on('resumeTokenChanged', (token) => {
      this.resumeTokens.set(collectionName, token);
      this.persistResumeToken(collectionName, token);
    });
  }

  private handleChange(collectionName: string, change: ChangeStreamDocument): void {
    switch (change.operationType) {
      case 'insert':
        this.handleInsert(collectionName, change);
        break;
      case 'update':
        this.handleUpdate(collectionName, change);
        break;
      case 'delete':
        this.handleDelete(collectionName, change);
        break;
      case 'invalidate':
        this.handleInvalidate(collectionName);
        break;
    }
  }

  private handleInsert(collectionName: string, change: ChangeStreamDocument): void {
    if (collectionName === 'photos' && change.fullDocument) {
      // Validate and sync new photo to mobile clients
      try {
        const photo = validatePhotoMetadata(change.fullDocument);
        this.syncToMobileClients('photo_created', photo);
      } catch (error) {
        console.error('Invalid photo document from change stream:', error);
      }
    }
  }

  private handleUpdate(collectionName: string, change: ChangeStreamDocument): void {
    if (collectionName === 'photos' && change.fullDocument) {
      try {
        const photo = validatePhotoMetadata(change.fullDocument);
        
        // Check what fields were updated
        const updatedFields = Object.keys(change.updateDescription?.updatedFields || {});
        
        this.syncToMobileClients('photo_updated', {
          photo,
          updatedFields,
          updateTimestamp: new Date()
        });
      } catch (error) {
        console.error('Invalid photo update from change stream:', error);
      }
    }
  }

  private handleDelete(collectionName: string, change: ChangeStreamDocument): void {
    this.syncToMobileClients('photo_deleted', {
      id: change.documentKey?._id,
      deletedAt: new Date()
    });
  }

  private handleInvalidate(collectionName: string): void {
    console.warn(`Change stream invalidated for collection: ${collectionName}`);
    // Restart the change stream
    this.restartChangeStream(collectionName);
  }

  private async restartChangeStream(collectionName: string): Promise<void> {
    const existingStream = this.changeStreams.get(collectionName);
    if (existingStream) {
      await existingStream.close();
      this.changeStreams.delete(collectionName);
    }

    // Restart with last known resume token
    const resumeToken = this.resumeTokens.get(collectionName);
    await this.setupChangeStream(collectionName, { resumeAfter: resumeToken });
  }

  private handleChangeStreamError(collectionName: string, error: any): void {
    if (error.code === 40585) {
      // Resume token expired - restart from current time
      console.warn(`Resume token expired for ${collectionName}, restarting from current time`);
      this.resumeTokens.delete(collectionName);
      this.restartChangeStream(collectionName);
    } else {
      console.error(`Unhandled change stream error for ${collectionName}:`, error);
    }
  }

  private syncToMobileClients(eventType: string, data: any): void {
    // Implement your mobile sync logic here
    // This could be WebSocket, Server-Sent Events, or push notifications
    console.log(`Syncing to mobile clients: ${eventType}`, data);
    
    // Example: Send to PowerSync or custom sync service
    this.notifyPowerSync(eventType, data);
  }

  private notifyPowerSync(eventType: string, data: any): void {
    // PowerSync integration
    // PowerSync will handle the actual sync to mobile clients
  }

  private async persistResumeToken(collectionName: string, token: ResumeToken): Promise<void> {
    // Persist resume token to survive service restarts
    await this.mongoService.getCollection('resume_tokens').updateOne(
      { collection: collectionName },
      {
        $set: {
          collection: collectionName,
          token: token,
          updatedAt: new Date()
        }
      },
      { upsert: true }
    );
  }

  async stopListening(): Promise<void> {
    if (!this.isListening) return;

    for (const [collectionName, stream] of this.changeStreams) {
      await stream.close();
      console.log(`Closed change stream for ${collectionName}`);
    }

    this.changeStreams.clear();
    this.isListening = false;
  }
}
```

### Custom Sync Service Architecture

```typescript
interface SyncEvent {
  id: string;
  type: 'create' | 'update' | 'delete';
  collection: string;
  documentId: string;
  timestamp: Date;
  data?: any;
  userId?: string;
  deviceId?: string;
}

export class CustomSyncService {
  private eventQueue: SyncEvent[] = [];
  private clientConnections = new Map<string, WebSocket>();
  private processingInterval?: NodeJS.Timer;

  constructor(
    private mongoService: MongoDBService,
    private changeStreamManager: MongoChangeStreamManager
  ) {
    this.setupEventProcessing();
  }

  private setupEventProcessing(): void {
    this.processingInterval = setInterval(() => {
      this.processEventQueue();
    }, 1000); // Process events every second
  }

  async queueSyncEvent(event: Omit<SyncEvent, 'id' | 'timestamp'>): Promise<void> {
    const syncEvent: SyncEvent = {
      ...event,
      id: this.generateEventId(),
      timestamp: new Date()
    };

    this.eventQueue.push(syncEvent);
    
    // Also persist to MongoDB for reliability
    await this.mongoService.getCollection('sync_events').insertOne(syncEvent);
  }

  private async processEventQueue(): Promise<void> {
    if (this.eventQueue.length === 0) return;

    const events = this.eventQueue.splice(0, 10); // Process 10 events at a time

    for (const event of events) {
      await this.processEvent(event);
    }
  }

  private async processEvent(event: SyncEvent): Promise<void> {
    // Send to all connected mobile clients
    const message = JSON.stringify({
      type: 'sync_event',
      event: event
    });

    for (const [clientId, connection] of this.clientConnections) {
      if (connection.readyState === WebSocket.OPEN) {
        try {
          connection.send(message);
        } catch (error) {
          console.error(`Failed to send sync event to client ${clientId}:`, error);
          this.clientConnections.delete(clientId);
        }
      }
    }

    // Send to PowerSync if configured
    await this.sendToPowerSync(event);
  }

  private async sendToPowerSync(event: SyncEvent): Promise<void> {
    // Implementation depends on your PowerSync setup
    // This would typically be an HTTP POST to PowerSync webhook
    try {
      await fetch(`${process.env.POWERSYNC_WEBHOOK_URL}/sync`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(event)
      });
    } catch (error) {
      console.error('Failed to send event to PowerSync:', error);
    }
  }

  addClient(clientId: string, connection: WebSocket): void {
    this.clientConnections.set(clientId, connection);
    
    connection.on('close', () => {
      this.clientConnections.delete(clientId);
    });

    // Send initial sync data to new client
    this.sendInitialSyncData(clientId, connection);
  }

  private async sendInitialSyncData(clientId: string, connection: WebSocket): Promise<void> {
    // Send recent changes to newly connected client
    const recentEvents = await this.mongoService.getCollection('sync_events')
      .find({
        timestamp: { $gte: new Date(Date.now() - 86400000) } // Last 24 hours
      })
      .sort({ timestamp: -1 })
      .limit(100)
      .toArray();

    const initialSyncMessage = JSON.stringify({
      type: 'initial_sync',
      events: recentEvents
    });

    if (connection.readyState === WebSocket.OPEN) {
      connection.send(initialSyncMessage);
    }
  }

  private generateEventId(): string {
    return `sync_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  async shutdown(): Promise<void> {
    if (this.processingInterval) {
      clearInterval(this.processingInterval);
    }

    // Close all client connections
    for (const connection of this.clientConnections.values()) {
      connection.close();
    }
    
    this.clientConnections.clear();
  }
}
```

---

## Performance Optimization Patterns

### MongoDB Indexing Strategies

```typescript
export class MongoIndexManager {
  constructor(private mongoService: MongoDBService) {}

  async createOptimalIndexes(): Promise<void> {
    await Promise.all([
      this.createPhotoIndexes(),
      this.createProjectIndexes(),
      this.createDepartmentIndexes(),
      this.createSyncIndexes()
    ]);
  }

  private async createPhotoIndexes(): Promise<void> {
    const photosCollection = this.mongoService.getCollection('photos');

    // Compound indexes for common query patterns
    const indexes = [
      // Department-based queries (most common)
      { departmentId: 1, capturedAt: -1 },
      
      // Project-based queries  
      { projectId: 1, capturedAt: -1 },
      
      // Sync status queries
      { syncStatus: 1, lastModified: -1 },
      
      // Search by part
      { partId: 1, capturedAt: -1 },
      
      // File management
      { fileName: 1 },
      { checksumMD5: 1 },
      
      // Geospatial index for location-based queries
      { 'metadata.location': '2dsphere' },
      
      // Text search index
      { 
        'metadata.inspectorNotes': 'text',
        'metadata.defectsFound': 'text',
        fileName: 'text'
      },
      
      // TTL index for cleanup (optional)
      { createdAt: 1 }, // Add expireAfterSeconds if needed
      
      // Compound index for complex queries
      {
        departmentId: 1,
        projectId: 1,
        syncStatus: 1,
        capturedAt: -1
      }
    ];

    for (const indexSpec of indexes) {
      try {
        await photosCollection.createIndex(indexSpec, {
          background: true, // Create index in background
          name: this.generateIndexName(indexSpec)
        });
      } catch (error) {
        console.error(`Failed to create index ${JSON.stringify(indexSpec)}:`, error);
      }
    }
  }

  private async createProjectIndexes(): Promise<void> {
    const projectsCollection = this.mongoService.getCollection('projects');

    const indexes = [
      { jobNumber: 1 }, // Unique constraint should be added separately
      { status: 1, startDate: -1 },
      { departments: 1 }, // Array index
      { client: 1, startDate: -1 },
      { expectedEndDate: 1 }
    ];

    for (const indexSpec of indexes) {
      await projectsCollection.createIndex(indexSpec, { background: true });
    }

    // Unique constraint for job numbers
    await projectsCollection.createIndex(
      { jobNumber: 1 },
      { unique: true, background: true }
    );
  }

  private async createSyncIndexes(): Promise<void> {
    const syncEventsCollection = this.mongoService.getCollection('sync_events');

    const indexes = [
      { timestamp: -1 }, // TTL index
      { collection: 1, timestamp: -1 },
      { documentId: 1, timestamp: -1 },
      { deviceId: 1, timestamp: -1 }
    ];

    for (const indexSpec of indexes) {
      await syncEventsCollection.createIndex(indexSpec, { background: true });
    }

    // TTL index to auto-delete old sync events
    await syncEventsCollection.createIndex(
      { timestamp: 1 },
      { 
        expireAfterSeconds: 7 * 24 * 60 * 60, // 7 days
        background: true 
      }
    );
  }

  private generateIndexName(indexSpec: any): string {
    return Object.entries(indexSpec)
      .map(([key, value]) => `${key}_${value}`)
      .join('_');
  }

  async analyzeSlowQueries(): Promise<SlowQueryAnalysis[]> {
    // Enable profiling for slow queries
    await this.mongoService.getCollection('photos').database.command({
      profile: 2,
      slowms: 100 // Queries taking longer than 100ms
    });

    // Wait for some queries to be profiled
    await new Promise(resolve => setTimeout(resolve, 60000)); // 1 minute

    // Retrieve profiling data
    const profilingData = await this.mongoService.getCollection('system.profile')
      .find({})
      .sort({ ts: -1 })
      .limit(100)
      .toArray();

    return profilingData.map(entry => ({
      command: entry.command,
      executionTimeMs: entry.millis,
      timestamp: entry.ts,
      planSummary: entry.planSummary,
      docsExamined: entry.docsExamined,
      docsReturned: entry.docsReturned
    }));
  }
}

interface SlowQueryAnalysis {
  command: any;
  executionTimeMs: number;
  timestamp: Date;
  planSummary?: string;
  docsExamined?: number;
  docsReturned?: number;
}
```

### Aggregation Pipeline Optimizations

```typescript
export class MongoAggregationService {
  constructor(private mongoService: MongoDBService) {}

  async getManufacturingDashboardData(
    departmentId?: DepartmentId,
    dateRange?: { start: Date; end: Date }
  ): Promise<DashboardData> {
    const photosCollection = this.mongoService.getCollection<PhotoMetadata>('photos');

    // Build match stage
    const matchStage: any = {};
    
    if (departmentId) {
      matchStage.departmentId = departmentId;
    }
    
    if (dateRange) {
      matchStage.capturedAt = {
        $gte: dateRange.start,
        $lte: dateRange.end
      };
    }

    // Optimized aggregation pipeline
    const pipeline = [
      // Match stage should be first for performance
      { $match: matchStage },
      
      // Add computed fields
      {
        $addFields: {
          capturedDate: {
            $dateToString: {
              format: '%Y-%m-%d',
              date: '$capturedAt'
            }
          },
          fileSizeMB: { $divide: ['$fileSize', 1048576] }
        }
      },
      
      // Facet for multiple aggregations in one query
      {
        $facet: {
          // Daily photo counts
          dailyStats: [
            {
              $group: {
                _id: '$capturedDate',
                photoCount: { $sum: 1 },
                totalSizeMB: { $sum: '$fileSizeMB' },
                avgQuality: { $avg: '$metadata.qualityRating' }
              }
            },
            { $sort: { _id: -1 } },
            { $limit: 30 } // Last 30 days
          ],
          
          // Department breakdown
          departmentStats: [
            {
              $group: {
                _id: '$departmentId',
                photoCount: { $sum: 1 },
                avgFileSize: { $avg: '$fileSize' },
                syncSuccessRate: {
                  $avg: {
                    $cond: [{ $eq: ['$syncStatus', 'synced'] }, 1, 0]
                  }
                }
              }
            }
          ],
          
          // Project breakdown
          projectStats: [
            {
              $group: {
                _id: '$projectId',
                photoCount: { $sum: 1 },
                latestPhoto: { $max: '$capturedAt' },
                defectCount: {
                  $sum: { $size: '$metadata.defectsFound' }
                }
              }
            },
            { $sort: { photoCount: -1 } },
            { $limit: 10 }
          ],
          
          // Quality metrics
          qualityStats: [
            {
              $group: {
                _id: '$metadata.qualityRating',
                count: { $sum: 1 },
                avgDefects: {
                  $avg: { $size: '$metadata.defectsFound' }
                }
              }
            },
            { $sort: { _id: 1 } }
          ],
          
          // Sync status
          syncStats: [
            {
              $group: {
                _id: '$syncStatus',
                count: { $sum: 1 },
                avgAttempts: { $avg: '$syncAttempts' }
              }
            }
          ],
          
          // Overall summary
          summary: [
            {
              $group: {
                _id: null,
                totalPhotos: { $sum: 1 },
                totalSizeMB: { $sum: '$fileSizeMB' },
                avgQuality: { $avg: '$metadata.qualityRating' },
                uniqueProjects: { $addToSet: '$projectId' },
                dateRange: {
                  $push: {
                    min: { $min: '$capturedAt' },
                    max: { $max: '$capturedAt' }
                  }
                }
              }
            },
            {
              $addFields: {
                uniqueProjectCount: { $size: '$uniqueProjects' }
              }
            }
          ]
        }
      }
    ];

    const [result] = await photosCollection.aggregate(pipeline).toArray();
    
    return {
      dailyStats: result.dailyStats,
      departmentStats: result.departmentStats,
      projectStats: result.projectStats,
      qualityStats: result.qualityStats,
      syncStats: result.syncStats,
      summary: result.summary[0],
      generatedAt: new Date()
    };
  }

  async getPhotoSearchResults(
    searchParams: PhotoSearchParams
  ): Promise<PhotoSearchResult> {
    const photosCollection = this.mongoService.getCollection<PhotoMetadata>('photos');

    // Build search pipeline
    const pipeline: any[] = [];

    // Text search stage (if text query provided)
    if (searchParams.textQuery) {
      pipeline.push({
        $match: {
          $text: {
            $search: searchParams.textQuery,
            $caseSensitive: false
          }
        }
      });
      
      // Add score for text search relevance
      pipeline.push({
        $addFields: {
          searchScore: { $meta: 'textScore' }
        }
      });
    }

    // Filter stage
    const filterStage: any = {};
    
    if (searchParams.departmentId) {
      filterStage.departmentId = searchParams.departmentId;
    }
    
    if (searchParams.projectId) {
      filterStage.projectId = searchParams.projectId;
    }
    
    if (searchParams.dateRange) {
      filterStage.capturedAt = {
        $gte: searchParams.dateRange.start,
        $lte: searchParams.dateRange.end
      };
    }
    
    if (searchParams.qualityRange) {
      filterStage['metadata.qualityRating'] = {
        $gte: searchParams.qualityRange.min,
        $lte: searchParams.qualityRange.max
      };
    }

    if (searchParams.hasDefects !== undefined) {
      if (searchParams.hasDefects) {
        filterStage['metadata.defectsFound'] = { $ne: [] };
      } else {
        filterStage['metadata.defectsFound'] = { $eq: [] };
      }
    }

    if (Object.keys(filterStage).length > 0) {
      pipeline.push({ $match: filterStage });
    }

    // Sort stage
    const sortStage: any = {};
    if (searchParams.textQuery) {
      sortStage.searchScore = { $meta: 'textScore' };
    }
    sortStage.capturedAt = -1; // Secondary sort by date

    pipeline.push({ $sort: sortStage });

    // Pagination
    const skip = (searchParams.page - 1) * searchParams.limit;
    pipeline.push({ $skip: skip });
    pipeline.push({ $limit: searchParams.limit });

    // Execute search
    const photos = await photosCollection.aggregate(pipeline).toArray();

    // Get total count for pagination
    const countPipeline = [...pipeline.slice(0, -2)]; // Remove skip and limit
    countPipeline.push({ $count: 'total' });
    const countResult = await photosCollection.aggregate(countPipeline).toArray();
    const totalCount = countResult[0]?.total || 0;

    return {
      photos: photos.map(p => validatePhotoMetadata(p)),
      totalCount,
      page: searchParams.page,
      totalPages: Math.ceil(totalCount / searchParams.limit),
      hasNext: skip + photos.length < totalCount,
      hasPrev: searchParams.page > 1
    };
  }
}

interface DashboardData {
  dailyStats: any[];
  departmentStats: any[];
  projectStats: any[];
  qualityStats: any[];
  syncStats: any[];
  summary: any;
  generatedAt: Date;
}

interface PhotoSearchParams {
  textQuery?: string;
  departmentId?: DepartmentId;
  projectId?: ProjectId;
  dateRange?: {
    start: Date;
    end: Date;
  };
  qualityRange?: {
    min: number;
    max: number;
  };
  hasDefects?: boolean;
  page: number;
  limit: number;
}

interface PhotoSearchResult {
  photos: PhotoMetadata[];
  totalCount: number;
  page: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}
```

---

## Data Migration and Schema Evolution

### Schema Versioning Strategy

```typescript
interface SchemaVersion {
  version: number;
  appliedAt: Date;
  description: string;
  migrationScript: string;
}

export class MongoMigrationManager {
  constructor(private mongoService: MongoDBService) {}

  async getCurrentSchemaVersion(): Promise<number> {
    const versionsCollection = this.mongoService.getCollection<SchemaVersion>('schema_versions');
    
    const latestVersion = await versionsCollection
      .findOne({}, { sort: { version: -1 } });
    
    return latestVersion?.version || 0;
  }

  async runMigrations(): Promise<void> {
    const currentVersion = await this.getCurrentSchemaVersion();
    const migrations = this.getMigrations();
    
    const pendingMigrations = migrations.filter(m => m.version > currentVersion);
    
    if (pendingMigrations.length === 0) {
      console.log('No pending migrations');
      return;
    }

    console.log(`Running ${pendingMigrations.length} migrations...`);

    for (const migration of pendingMigrations) {
      await this.runMigration(migration);
    }

    console.log('All migrations completed successfully');
  }

  private getMigrations(): Migration[] {
    return [
      {
        version: 1,
        description: 'Add sync status and attempts fields',
        up: async (db) => {
          await db.collection('photos').updateMany(
            { syncStatus: { $exists: false } },
            {
              $set: {
                syncStatus: 'pending',
                syncAttempts: 0,
                lastModified: new Date()
              }
            }
          );
        }
      },
      {
        version: 2,
        description: 'Add department-specific settings',
        up: async (db) => {
          await db.collection('departments').updateMany(
            { settings: { $exists: false } },
            {
              $set: {
                settings: {
                  maxPhotosPerSession: 100,
                  compressionQuality: 90,
                  requireGPSLocation: false,
                  mandatoryFields: [],
                  photoRetentionDays: 365
                }
              }
            }
          );
        }
      },
      {
        version: 3,
        description: 'Add audit trail to photos',
        up: async (db) => {
          const photosCollection = db.collection('photos');
          const photos = await photosCollection.find({
            auditTrail: { $exists: false }
          }).toArray();

          for (const photo of photos) {
            await photosCollection.updateOne(
              { _id: photo._id },
              {
                $set: {
                  auditTrail: [{
                    action: 'created',
                    timestamp: photo.capturedAt || photo.createdAt,
                    userId: photo.metadata?.deviceInfo?.deviceId || 'unknown',
                    deviceId: photo.metadata?.deviceInfo?.deviceId || 'unknown',
                    details: {
                      migrated: true,
                      originalCreatedAt: photo.createdAt
                    }
                  }]
                }
              }
            );
          }
        }
      },
      {
        version: 4,
        description: 'Create compound indexes for performance',
        up: async (db) => {
          const photosCollection = db.collection('photos');
          
          await photosCollection.createIndex(
            { departmentId: 1, capturedAt: -1 },
            { background: true }
          );
          
          await photosCollection.createIndex(
            { projectId: 1, syncStatus: 1 },
            { background: true }
          );
        }
      }
    ];
  }

  private async runMigration(migration: Migration): Promise<void> {
    const session = this.mongoService.client.startSession();
    
    try {
      await session.withTransaction(async () => {
        console.log(`Running migration ${migration.version}: ${migration.description}`);
        
        await migration.up(this.mongoService.db);
        
        // Record the migration
        await this.mongoService.getCollection<SchemaVersion>('schema_versions').insertOne({
          version: migration.version,
          appliedAt: new Date(),
          description: migration.description,
          migrationScript: migration.up.toString()
        });
      });
      
      console.log(`Migration ${migration.version} completed successfully`);
    } catch (error) {
      console.error(`Migration ${migration.version} failed:`, error);
      throw error;
    } finally {
      await session.endSession();
    }
  }

  async rollback(targetVersion: number): Promise<void> {
    const currentVersion = await this.getCurrentSchemaVersion();
    
    if (targetVersion >= currentVersion) {
      console.log('Target version is not lower than current version');
      return;
    }

    const migrations = this.getMigrations();
    const rollbackMigrations = migrations
      .filter(m => m.version > targetVersion && m.down)
      .reverse(); // Rollback in reverse order

    for (const migration of rollbackMigrations) {
      console.log(`Rolling back migration ${migration.version}`);
      
      const session = this.mongoService.client.startSession();
      
      try {
        await session.withTransaction(async () => {
          if (migration.down) {
            await migration.down(this.mongoService.db);
          }
          
          // Remove migration record
          await this.mongoService.getCollection('schema_versions').deleteOne({
            version: migration.version
          });
        });
      } finally {
        await session.endSession();
      }
    }

    console.log(`Rollback to version ${targetVersion} completed`);
  }
}

interface Migration {
  version: number;
  description: string;
  up: (db: Db) => Promise<void>;
  down?: (db: Db) => Promise<void>;
}
```

---

## Security Best Practices

### Authentication and Authorization

```typescript
import jwt from 'jsonwebtoken';
import bcrypt from 'bcrypt';

interface DeviceToken {
  deviceId: string;
  departmentId: DepartmentId;
  permissions: string[];
  issuedAt: number;
  expiresAt: number;
}

export class MongoSecurityService {
  constructor(
    private mongoService: MongoDBService,
    private jwtSecret: string
  ) {}

  async authenticateDevice(
    deviceId: string,
    departmentCode: string,
    accessPin: string
  ): Promise<AuthResult> {
    // Verify department and access pin
    const department = await this.mongoService.getCollection<Department>('departments')
      .findOne({ code: departmentCode });

    if (!department || !department.active) {
      return { success: false, error: 'Invalid department' };
    }

    // In a real implementation, you'd verify the access pin
    // This is a simplified example
    const validPin = await this.verifyAccessPin(departmentCode, accessPin);
    if (!validPin) {
      return { success: false, error: 'Invalid access pin' };
    }

    // Generate JWT token
    const tokenData: DeviceToken = {
      deviceId,
      departmentId: department._id as any,
      permissions: this.getDepartmentPermissions(department),
      issuedAt: Date.now(),
      expiresAt: Date.now() + (24 * 60 * 60 * 1000) // 24 hours
    };

    const token = jwt.sign(tokenData, this.jwtSecret, { expiresIn: '24h' });

    // Store device session
    await this.mongoService.getCollection('device_sessions').insertOne({
      deviceId,
      departmentId: department._id,
      token,
      createdAt: new Date(),
      expiresAt: new Date(tokenData.expiresAt),
      lastActivity: new Date()
    });

    return {
      success: true,
      token,
      department: validateDepartment(department),
      permissions: tokenData.permissions
    };
  }

  async verifyToken(token: string): Promise<TokenVerification> {
    try {
      const decoded = jwt.verify(token, this.jwtSecret) as DeviceToken;
      
      // Check if session is still active
      const session = await this.mongoService.getCollection('device_sessions')
        .findOne({
          deviceId: decoded.deviceId,
          token,
          expiresAt: { $gt: new Date() }
        });

      if (!session) {
        return { valid: false, error: 'Session expired or invalid' };
      }

      // Update last activity
      await this.mongoService.getCollection('device_sessions').updateOne(
        { _id: session._id },
        { $set: { lastActivity: new Date() } }
      );

      return {
        valid: true,
        deviceToken: decoded,
        sessionId: session._id.toString()
      };
    } catch (error) {
      return { valid: false, error: 'Invalid token' };
    }
  }

  async authorizeOperation(
    token: string,
    operation: string,
    resource: any
  ): Promise<boolean> {
    const verification = await this.verifyToken(token);
    
    if (!verification.valid || !verification.deviceToken) {
      return false;
    }

    const { deviceToken } = verification;

    // Check department-based authorization
    if (resource.departmentId && resource.departmentId !== deviceToken.departmentId) {
      // Device can only access its own department's data
      return false;
    }

    // Check operation permissions
    const requiredPermission = this.getRequiredPermission(operation);
    if (!deviceToken.permissions.includes(requiredPermission)) {
      return false;
    }

    return true;
  }

  private async verifyAccessPin(departmentCode: string, pin: string): Promise<boolean> {
    // In a real implementation, you'd have hashed pins stored securely
    const departmentAuth = await this.mongoService.getCollection('department_auth')
      .findOne({ departmentCode });

    if (!departmentAuth) return false;

    return await bcrypt.compare(pin, departmentAuth.hashedPin);
  }

  private getDepartmentPermissions(department: Department): string[] {
    const basePermissions = ['photo:create', 'photo:read'];
    
    if (department.name === 'QA/QC') {
      basePermissions.push('photo:update', 'photo:delete', 'quality:review');
    }

    return basePermissions;
  }

  private getRequiredPermission(operation: string): string {
    const permissionMap: Record<string, string> = {
      'create_photo': 'photo:create',
      'read_photo': 'photo:read',
      'update_photo': 'photo:update',
      'delete_photo': 'photo:delete',
      'review_quality': 'quality:review'
    };

    return permissionMap[operation] || 'general:access';
  }
}

interface AuthResult {
  success: boolean;
  token?: string;
  department?: Department;
  permissions?: string[];
  error?: string;
}

interface TokenVerification {
  valid: boolean;
  deviceToken?: DeviceToken;
  sessionId?: string;
  error?: string;
}
```

### Data Encryption and Privacy

```typescript
import crypto from 'crypto';

export class MongoEncryptionService {
  private readonly algorithm = 'aes-256-gcm';
  private readonly keyLength = 32;
  private readonly ivLength = 16;

  constructor(private encryptionKey: Buffer) {
    if (encryptionKey.length !== this.keyLength) {
      throw new Error(`Encryption key must be ${this.keyLength} bytes`);
    }
  }

  encryptSensitiveData(data: any): EncryptedData {
    const iv = crypto.randomBytes(this.ivLength);
    const cipher = crypto.createCipher(this.algorithm, this.encryptionKey);
    cipher.setAAD(Buffer.from('manufacturing-photos'));

    const jsonData = JSON.stringify(data);
    
    let encrypted = cipher.update(jsonData, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    
    const authTag = cipher.getAuthTag();

    return {
      encrypted,
      iv: iv.toString('hex'),
      authTag: authTag.toString('hex'),
      algorithm: this.algorithm
    };
  }

  decryptSensitiveData(encryptedData: EncryptedData): any {
    const decipher = crypto.createDecipher(
      encryptedData.algorithm,
      this.encryptionKey
    );
    
    decipher.setAAD(Buffer.from('manufacturing-photos'));
    decipher.setAuthTag(Buffer.from(encryptedData.authTag, 'hex'));

    let decrypted = decipher.update(encryptedData.encrypted, 'hex', 'utf8');
    decrypted += decipher.final('utf8');

    return JSON.parse(decrypted);
  }

  // Field-level encryption for specific sensitive fields
  async encryptPhotoMetadata(metadata: PhotoMetadata): Promise<PhotoMetadata> {
    const encryptedMetadata = { ...metadata };

    // Encrypt inspector notes if present
    if (metadata.metadata.inspectorNotes) {
      const encrypted = this.encryptSensitiveData(metadata.metadata.inspectorNotes);
      encryptedMetadata.metadata = {
        ...encryptedMetadata.metadata,
        inspectorNotesEncrypted: encrypted
      };
      delete encryptedMetadata.metadata.inspectorNotes;
    }

    // Encrypt GPS location for privacy
    if (metadata.metadata.location) {
      const encrypted = this.encryptSensitiveData(metadata.metadata.location);
      encryptedMetadata.metadata = {
        ...encryptedMetadata.metadata,
        locationEncrypted: encrypted
      };
      delete encryptedMetadata.metadata.location;
    }

    return encryptedMetadata;
  }

  async decryptPhotoMetadata(encryptedMetadata: any): Promise<PhotoMetadata> {
    const decryptedMetadata = { ...encryptedMetadata };

    // Decrypt inspector notes
    if (encryptedMetadata.metadata.inspectorNotesEncrypted) {
      decryptedMetadata.metadata.inspectorNotes = this.decryptSensitiveData(
        encryptedMetadata.metadata.inspectorNotesEncrypted
      );
      delete decryptedMetadata.metadata.inspectorNotesEncrypted;
    }

    // Decrypt GPS location
    if (encryptedMetadata.metadata.locationEncrypted) {
      decryptedMetadata.metadata.location = this.decryptSensitiveData(
        encryptedMetadata.metadata.locationEncrypted
      );
      delete decryptedMetadata.metadata.locationEncrypted;
    }

    return validatePhotoMetadata(decryptedMetadata);
  }
}

interface EncryptedData {
  encrypted: string;
  iv: string;
  authTag: string;
  algorithm: string;
}
```

---

## Testing MongoDB Integration

### Integration Testing

```typescript
import { MongoMemoryServer } from 'mongodb-memory-server';
import { MongoClient } from 'mongodb';

describe('MongoDB Integration Tests', () => {
  let mongoServer: MongoMemoryServer;
  let mongoClient: MongoClient;
  let mongoService: MongoDBService;
  let photoRepository: PhotoRepository;

  beforeAll(async () => {
    // Start in-memory MongoDB server
    mongoServer = await MongoMemoryServer.create();
    const uri = mongoServer.getUri();
    
    mongoClient = new MongoClient(uri);
    await mongoClient.connect();

    mongoService = new MongoDBService({
      connectionString: uri,
      databaseName: 'test_manufacturing'
    });
    
    await mongoService.connect();
    
    photoRepository = new PhotoRepository(
      mongoService.getCollection('photos')
    );
  });

  afterAll(async () => {
    await mongoClient.close();
    await mongoService.disconnect();
    await mongoServer.stop();
  });

  beforeEach(async () => {
    // Clean up collections before each test
    await mongoService.getCollection('photos').deleteMany({});
    await mongoService.getCollection('departments').deleteMany({});
    await mongoService.getCollection('projects').deleteMany({});
  });

  test('should create and retrieve photo with validation', async () => {
    const photoData = {
      localPath: '/test/photo.jpg',
      fileName: 'test-photo.jpg',
      departmentId: 'PAINT' as DepartmentId,
      projectId: new ObjectId() as ProjectId,
      partId: 'PART-001' as PartId,
      jobNumber: 'JOB-2025-001',
      capturedAt: new Date(),
      lastModified: new Date(),
      fileSize: 1048576,
      mimeType: 'image/jpeg',
      checksumMD5: '5d41402abc4b2a76b9719d911017c592',
      syncStatus: 'pending' as const,
      syncAttempts: 0,
      metadata: {
        inspectionType: 'quality-check',
        qualityRating: 8,
        defectsFound: [],
        deviceInfo: {
          deviceId: 'device-001',
          model: 'Samsung Galaxy',
          osVersion: 'Android 12',
          appVersion: '1.0.0',
          cameraSettings: {
            quality: 100,
            flash: 'auto' as const,
            resolution: { width: 2048, height: 2048 }
          }
        }
      },
      auditTrail: []
    };

    const createdPhoto = await photoRepository.create(photoData);

    expect(createdPhoto._id).toBeDefined();
    expect(createdPhoto.departmentId).toBe('PAINT');
    expect(createdPhoto.syncStatus).toBe('pending');

    const retrievedPhoto = await photoRepository.findById(createdPhoto._id);
    expect(retrievedPhoto).toBeDefined();
    expect(retrievedPhoto?.fileName).toBe('test-photo.jpg');
  });

  test('should handle validation errors gracefully', async () => {
    const invalidPhotoData = {
      localPath: '/test/photo.jpg',
      fileName: '', // Invalid: empty filename
      departmentId: 'INVALID_DEPT', // Invalid department
      // Missing required fields
    };

    await expect(
      photoRepository.create(invalidPhotoData as any)
    ).rejects.toThrow();
  });

  test('should perform complex aggregation queries', async () => {
    // Insert test data
    const testPhotos = [
      {
        departmentId: 'PAINT' as DepartmentId,
        projectId: new ObjectId() as ProjectId,
        capturedAt: new Date('2025-01-01'),
        metadata: { qualityRating: 8, defectsFound: [] }
      },
      {
        departmentId: 'PAINT' as DepartmentId,  
        projectId: new ObjectId() as ProjectId,
        capturedAt: new Date('2025-01-02'),
        metadata: { qualityRating: 7, defectsFound: ['scratch'] }
      },
      {
        departmentId: 'QA' as DepartmentId,
        projectId: new ObjectId() as ProjectId,
        capturedAt: new Date('2025-01-03'),
        metadata: { qualityRating: 9, defectsFound: [] }
      }
    ].map(data => ({
      ...data,
      localPath: '/test/photo.jpg',
      fileName: 'test.jpg',
      partId: 'PART-001' as PartId,
      jobNumber: 'JOB-001',
      lastModified: new Date(),
      fileSize: 1000000,
      mimeType: 'image/jpeg',
      checksumMD5: '5d41402abc4b2a76b9719d911017c592',
      syncStatus: 'pending' as const,
      syncAttempts: 0,
      auditTrail: []
    }));

    for (const photo of testPhotos) {
      await photoRepository.create(photo);
    }

    // Test aggregation service
    const aggregationService = new MongoAggregationService(mongoService);
    
    const dashboardData = await aggregationService.getManufacturingDashboardData();

    expect(dashboardData.summary.totalPhotos).toBe(3);
    expect(dashboardData.departmentStats).toHaveLength(2);
    expect(dashboardData.qualityStats.some(s => s._id === 8)).toBe(true);
  });

  test('should handle concurrent operations safely', async () => {
    const photoData = {
      localPath: '/test/concurrent.jpg',
      fileName: 'concurrent-test.jpg',
      departmentId: 'PAINT' as DepartmentId,
      projectId: new ObjectId() as ProjectId,
      partId: 'PART-001' as PartId,
      jobNumber: 'JOB-CONCURRENT',
      capturedAt: new Date(),
      lastModified: new Date(),
      fileSize: 1048576,
      mimeType: 'image/jpeg',
      checksumMD5: '5d41402abc4b2a76b9719d911017c592',
      syncStatus: 'pending' as const,
      syncAttempts: 0,
      metadata: {
        inspectionType: 'concurrent-test',
        deviceInfo: {
          deviceId: 'device-001',
          model: 'Test Device',
          osVersion: 'Test OS',
          appVersion: '1.0.0',
          cameraSettings: {
            quality: 100,
            flash: 'auto' as const,
            resolution: { width: 2048, height: 2048 }
          }
        }
      },
      auditTrail: []
    };

    const photo = await photoRepository.create(photoData);

    // Simulate concurrent updates
    const updatePromises = Array.from({ length: 5 }, (_, i) =>
      photoRepository.update(photo._id, {
        $set: { 
          syncAttempts: i + 1,
          lastModified: new Date()
        }
      })
    );

    const results = await Promise.allSettled(updatePromises);
    
    // All updates should succeed due to MongoDB's document-level locking
    const successfulUpdates = results.filter(r => r.status === 'fulfilled');
    expect(successfulUpdates).toHaveLength(5);

    const finalPhoto = await photoRepository.findById(photo._id);
    expect(finalPhoto?.syncAttempts).toBeGreaterThan(0);
  });
});
```

### Performance Testing

```typescript
describe('MongoDB Performance Tests', () => {
  let mongoService: MongoDBService;
  let photoRepository: PhotoRepository;

  beforeAll(async () => {
    // Use actual MongoDB instance for performance testing
    mongoService = new MongoDBService({
      connectionString: process.env.MONGODB_TEST_URL!,
      databaseName: 'performance_test'
    });
    
    await mongoService.connect();
    photoRepository = new PhotoRepository(mongoService.getCollection('photos'));
  });

  afterAll(async () => {
    await mongoService.disconnect();
  });

  test('should handle large batch inserts efficiently', async () => {
    const batchSize = 1000;
    const photos = Array.from({ length: batchSize }, (_, i) => ({
      localPath: `/test/photo-${i}.jpg`,
      fileName: `photo-${i}.jpg`,
      departmentId: 'PAINT' as DepartmentId,
      projectId: new ObjectId() as ProjectId,
      partId: `PART-${String(i).padStart(3, '0')}` as PartId,
      jobNumber: 'PERF-TEST-001',
      capturedAt: new Date(),
      lastModified: new Date(),
      fileSize: Math.floor(Math.random() * 5000000) + 1000000, // 1-6MB
      mimeType: 'image/jpeg',
      checksumMD5: crypto.randomBytes(16).toString('hex'),
      syncStatus: 'pending' as const,
      syncAttempts: 0,
      metadata: {
        inspectionType: 'performance-test',
        qualityRating: Math.floor(Math.random() * 10) + 1,
        defectsFound: Math.random() > 0.7 ? ['test-defect'] : [],
        deviceInfo: {
          deviceId: `device-${i % 10}`,
          model: 'Performance Test Device',
          osVersion: 'Test OS 1.0',
          appVersion: '1.0.0',
          cameraSettings: {
            quality: 100,
            flash: 'auto' as const,
            resolution: { width: 2048, height: 2048 }
          }
        }
      },
      auditTrail: []
    }));

    const startTime = Date.now();

    // Batch insert using MongoDB's bulk operations
    await mongoService.getCollection('photos').insertMany(photos);

    const insertTime = Date.now() - startTime;
    console.log(`Inserted ${batchSize} photos in ${insertTime}ms`);

    // Performance expectations
    expect(insertTime).toBeLessThan(5000); // Should complete in under 5 seconds
    
    // Verify all photos were inserted
    const count = await photoRepository.count({ jobNumber: 'PERF-TEST-001' });
    expect(count).toBe(batchSize);
  });

  test('should query large datasets efficiently', async () => {
    const startTime = Date.now();

    // Test complex query with aggregation
    const aggregationService = new MongoAggregationService(mongoService);
    const dashboardData = await aggregationService.getManufacturingDashboardData(
      'PAINT' as DepartmentId,
      {
        start: new Date('2025-01-01'),
        end: new Date('2025-12-31')
      }
    );

    const queryTime = Date.now() - startTime;
    console.log(`Complex aggregation completed in ${queryTime}ms`);

    expect(queryTime).toBeLessThan(2000); // Should complete in under 2 seconds
    expect(dashboardData.summary).toBeDefined();
  });

  test('should handle concurrent read/write operations', async () => {
    const concurrentOperations = 50;
    const startTime = Date.now();

    const operations = Array.from({ length: concurrentOperations }, async (_, i) => {
      if (i % 2 === 0) {
        // Read operation
        return await photoRepository.findByDepartment('PAINT' as DepartmentId);
      } else {
        // Write operation
        return await photoRepository.create({
          localPath: `/test/concurrent-${i}.jpg`,
          fileName: `concurrent-${i}.jpg`,
          departmentId: 'PAINT' as DepartmentId,
          projectId: new ObjectId() as ProjectId,
          partId: `PART-CONCURRENT-${i}` as PartId,
          jobNumber: 'CONCURRENT-TEST',
          capturedAt: new Date(),
          lastModified: new Date(),
          fileSize: 1000000,
          mimeType: 'image/jpeg',
          checksumMD5: crypto.randomBytes(16).toString('hex'),
          syncStatus: 'pending' as const,
          syncAttempts: 0,
          metadata: {
            inspectionType: 'concurrent-test',
            deviceInfo: {
              deviceId: `concurrent-device-${i}`,
              model: 'Concurrent Test',
              osVersion: 'Test OS',
              appVersion: '1.0.0',
              cameraSettings: {
                quality: 100,
                flash: 'auto' as const,
                resolution: { width: 2048, height: 2048 }
              }
            }
          },
          auditTrail: []
        });
      }
    });

    const results = await Promise.allSettled(operations);
    const operationTime = Date.now() - startTime;

    console.log(`${concurrentOperations} concurrent operations completed in ${operationTime}ms`);

    const successfulOperations = results.filter(r => r.status === 'fulfilled').length;
    expect(successfulOperations).toBe(concurrentOperations);
    expect(operationTime).toBeLessThan(10000); // Should complete in under 10 seconds
  });
});
```

This comprehensive guide provides production-ready patterns for integrating MongoDB with TypeScript mobile applications, covering everything from basic CRUD operations to advanced real-time synchronization and performance optimization strategies.
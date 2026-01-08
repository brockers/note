# TypeScript Mobile Development Best Practices (2024-2025)

**Comprehensive guide for building production-ready TypeScript mobile applications with modern frameworks and offline-first architectures.**

---

## Framework Selection Guide

### React Native with TypeScript (Recommended for Most Cases)

**Current Status**: Most mature and widely adopted TypeScript mobile solution
- **Community**: Largest ecosystem with 100K+ GitHub stars
- **TypeScript Support**: First-class integration with comprehensive type definitions
- **Performance**: New Architecture (Fabric + TurboModules) provides native-level performance
- **Libraries**: Extensive plugin ecosystem with most maintained TypeScript definitions

**When to Choose React Native**:
- Need maximum ecosystem support and third-party libraries
- Team has React/JavaScript experience
- Require native-level performance for camera, database, or file operations
- Building complex offline-first applications
- Need proven production patterns and extensive documentation

**Key Libraries for TypeScript**:
```bash
# Essential TypeScript-compatible libraries
npm install react-native-vision-camera    # Camera with full TS support
npm install @nozbe/watermelondb           # Offline-first database
npm install @tanstack/react-query         # Network state management
npm install react-hook-form               # Type-safe form handling
npm install zod                          # Runtime schema validation
```

### Capacitor with TypeScript (Best for Web Teams)

**Current Status**: Modern hybrid framework with excellent TypeScript integration
- **TypeScript First**: Built with TypeScript from ground up
- **Web Compatibility**: Single codebase works as PWA and mobile app
- **Native Bridge**: Unified API for native functionality across platforms

**When to Choose Capacitor**:
- Team has strong web development background
- Need PWA compatibility alongside mobile
- Prefer simpler build/deployment process
- Want modern architecture without React Native complexity

**TypeScript Configuration**:
```typescript
// capacitor.config.ts
import { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.company.app',
  appName: 'TypeScript App',
  webDir: 'dist',
  bundledWebRuntime: false,
  plugins: {
    Camera: {
      permissions: ['camera', 'photos']
    }
  }
};

export default config;
```

### Expo with TypeScript (Fastest Development)

**Current Status**: Positioned as the preferred React Native solution for 2025
- **Managed Workflow**: Expo handles native code through Config Plugins
- **EAS Services**: Cloud build and deployment services
- **TypeScript Templates**: Ready-to-use TypeScript project templates

**When to Choose Expo**:
- Need fastest time to market
- Prefer managed infrastructure and services
- Team new to React Native development
- Building standard mobile app without complex native requirements

### Framework Decision Matrix

| Feature | React Native | Capacitor | Expo | NativeScript |
|---------|--------------|-----------|------|--------------|
| TypeScript Support | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| Native Performance | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Ecosystem Size | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ |
| Learning Curve | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| Offline-First | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |

---

## TypeScript Configuration Best Practices

### Strict TypeScript Setup

**Essential Compiler Options**:
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM"],
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "allowJs": false,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "node",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "react-native"
  }
}
```

### Branded Types for Mobile Apps

```typescript
// Create domain-specific types for better type safety
const UserIdSchema = z.string().uuid().brand<'UserId'>();
const DeviceIdSchema = z.string().brand<'DeviceId'>();
const PhotoIdSchema = z.string().uuid().brand<'PhotoId'>();

export type UserId = z.infer<typeof UserIdSchema>;
export type DeviceId = z.infer<typeof DeviceIdSchema>;
export type PhotoId = z.infer<typeof PhotoIdSchema>;

// Type guards prevent mixing incompatible IDs
function getUser(id: UserId): Promise<User> { /* ... */ }
function getDevice(id: DeviceId): Promise<Device> { /* ... */ }

// Compile-time safety
const userId = UserIdSchema.parse('user-123');
const deviceId = DeviceIdSchema.parse('device-456');

getUser(userId);    // ✅ Correct
getUser(deviceId);  // ❌ TypeScript error
```

### Mobile-Specific Type Patterns

```typescript
// Network state typing
type NetworkState = {
  isConnected: boolean;
  type: 'wifi' | 'cellular' | 'none';
  isInternetReachable?: boolean;
};

// Photo metadata with proper typing
interface PhotoMetadata {
  readonly id: PhotoId;
  readonly capturedAt: Date;
  readonly location?: {
    readonly latitude: number;
    readonly longitude: number;
  };
  readonly deviceInfo: {
    readonly model: string;
    readonly os: string;
    readonly appVersion: string;
  };
  syncStatus: 'pending' | 'syncing' | 'synced' | 'failed';
}

// Type-safe error handling
type Result<T, E = Error> = 
  | { success: true; data: T }
  | { success: false; error: E };

async function capturePhoto(): Promise<Result<PhotoMetadata, CameraError>> {
  try {
    const photo = await camera.takePhoto();
    return { success: true, data: photo };
  } catch (error) {
    return { success: false, error: error as CameraError };
  }
}
```

---

## Camera Integration Best Practices

### React Native Vision Camera (Recommended)

**Installation and Setup**:
```bash
npm install react-native-vision-camera react-native-permissions
```

**TypeScript Implementation**:
```typescript
import React, { useRef, useState, useEffect } from 'react';
import { Camera, useCameraDevice, PhotoFile, CameraPermissionStatus } from 'react-native-vision-camera';

interface CameraServiceConfig {
  quality: number;
  flash: 'auto' | 'on' | 'off';
  enableShutterSound: boolean;
}

export class TypeSafeCameraService {
  private camera = useRef<Camera>(null);
  private device = useCameraDevice('back');

  async requestPermissions(): Promise<boolean> {
    const permission = await Camera.requestCameraPermission();
    return permission === 'granted';
  }

  async capturePhoto(config: CameraServiceConfig): Promise<PhotoFile | null> {
    if (!this.camera.current || !this.device) {
      throw new Error('Camera not initialized');
    }

    try {
      const photo = await this.camera.current.takePhoto({
        quality: config.quality,
        flash: config.flash,
        enableShutterSound: config.enableShutterSound
      });

      return photo;
    } catch (error) {
      console.error('Photo capture failed:', error);
      return null;
    }
  }
}

// React Hook for camera operations
export function useCamera() {
  const [hasPermission, setHasPermission] = useState<boolean | null>(null);
  const [cameraService] = useState(() => new TypeSafeCameraService());

  useEffect(() => {
    cameraService.requestPermissions().then(setHasPermission);
  }, [cameraService]);

  return {
    hasPermission,
    capturePhoto: cameraService.capturePhoto.bind(cameraService)
  };
}
```

### Photo Processing with TypeScript

```typescript
import { manipulateAsync, SaveFormat, ImageResult } from 'expo-image-manipulator';

interface PhotoProcessingOptions {
  quality: number;
  maxWidth?: number;
  maxHeight?: number;
  format: 'jpeg' | 'png';
}

export class PhotoProcessor {
  static async compressPhoto(
    uri: string, 
    options: PhotoProcessingOptions
  ): Promise<ImageResult> {
    const actions = [];
    
    if (options.maxWidth || options.maxHeight) {
      actions.push({
        resize: {
          width: options.maxWidth,
          height: options.maxHeight,
        }
      });
    }

    return await manipulateAsync(
      uri,
      actions,
      {
        compress: options.quality,
        format: options.format === 'jpeg' ? SaveFormat.JPEG : SaveFormat.PNG,
      }
    );
  }

  static async generateThumbnail(uri: string): Promise<string> {
    const result = await this.compressPhoto(uri, {
      quality: 0.7,
      maxWidth: 300,
      maxHeight: 300,
      format: 'jpeg'
    });
    
    return result.uri;
  }
}
```

---

## Offline-First Architecture Patterns

### WatermelonDB with TypeScript

**Schema Definition**:
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
        { name: 'metadata_json', type: 'string' },
        { name: 'captured_at', type: 'number' },
        { name: 'sync_status', type: 'string' },
        { name: 'file_size', type: 'number' }
      ]
    }),
    tableSchema({
      name: 'sync_queue',
      columns: [
        { name: 'item_id', type: 'string' },
        { name: 'item_type', type: 'string' },
        { name: 'attempts', type: 'number' },
        { name: 'last_attempt', type: 'number', isOptional: true }
      ]
    })
  ]
});
```

**Model Implementation**:
```typescript
import { Model, field, date, readonly } from '@nozbe/watermelondb/decorators';
import { PhotoMetadata } from '../types';

export class Photo extends Model {
  static table = 'photos';

  @field('local_path') localPath!: string;
  @field('remote_url') remoteUrl?: string;
  @field('metadata_json') private metadataJson!: string;
  @date('captured_at') capturedAt!: Date;
  @field('sync_status') syncStatus!: 'pending' | 'synced' | 'failed';
  @readonly @date('created_at') createdAt!: Date;

  get metadata(): PhotoMetadata {
    return JSON.parse(this.metadataJson);
  }

  set metadata(value: PhotoMetadata) {
    this.metadataJson = JSON.stringify(value);
  }

  get isPendingSync(): boolean {
    return this.syncStatus === 'pending';
  }
}
```

### Repository Pattern with TypeScript

```typescript
interface Repository<T, K = string> {
  create(data: Omit<T, 'id'>): Promise<T>;
  findById(id: K): Promise<T | null>;
  findMany(criteria: QueryCriteria<T>): Promise<T[]>;
  update(id: K, updates: Partial<T>): Promise<T>;
  delete(id: K): Promise<boolean>;
}

export class PhotoRepository implements Repository<Photo> {
  constructor(private database: Database) {}

  async create(data: Omit<PhotoMetadata, 'id'>): Promise<Photo> {
    return await this.database.write(async () => {
      return await this.database.get<Photo>('photos').create(photo => {
        photo.localPath = data.localPath;
        photo.metadata = data.metadata;
        photo.capturedAt = data.capturedAt;
        photo.syncStatus = 'pending';
      });
    });
  }

  async findPendingSync(): Promise<Photo[]> {
    return await this.database.get<Photo>('photos')
      .query(Q.where('sync_status', 'pending'))
      .fetch();
  }

  async markAsSynced(id: string, remoteUrl: string): Promise<void> {
    const photo = await this.database.get<Photo>('photos').find(id);
    
    await this.database.write(async () => {
      await photo.update(photo => {
        photo.syncStatus = 'synced';
        photo.remoteUrl = remoteUrl;
      });
    });
  }
}
```

### Sync Manager with TypeScript

```typescript
interface SyncTask {
  id: string;
  type: 'photo_upload' | 'metadata_sync';
  data: unknown;
  attempts: number;
  nextAttempt: Date;
}

export class TypeSafeSyncManager {
  private syncQueue: Map<string, SyncTask> = new Map();
  private isOnline = false;
  private syncInProgress = false;

  constructor(
    private photoRepo: PhotoRepository,
    private networkMonitor: NetworkMonitor
  ) {
    this.networkMonitor.onStatusChange((status) => {
      this.isOnline = status.isConnected;
      if (this.isOnline) {
        this.processSyncQueue();
      }
    });
  }

  async queuePhotoUpload(photo: Photo): Promise<void> {
    const task: SyncTask = {
      id: photo.id,
      type: 'photo_upload',
      data: photo,
      attempts: 0,
      nextAttempt: new Date()
    };

    this.syncQueue.set(task.id, task);
    
    if (this.isOnline && !this.syncInProgress) {
      this.processSyncQueue();
    }
  }

  private async processSyncQueue(): Promise<void> {
    if (this.syncInProgress) return;
    
    this.syncInProgress = true;
    const now = new Date();
    
    try {
      const readyTasks = Array.from(this.syncQueue.values())
        .filter(task => task.nextAttempt <= now)
        .slice(0, 5); // Process 5 at a time

      for (const task of readyTasks) {
        try {
          await this.executeTask(task);
          this.syncQueue.delete(task.id);
        } catch (error) {
          await this.handleTaskFailure(task, error);
        }
      }
    } finally {
      this.syncInProgress = false;
    }
  }

  private async handleTaskFailure(task: SyncTask, error: unknown): Promise<void> {
    task.attempts++;
    
    if (task.attempts >= 5) {
      this.syncQueue.delete(task.id);
      console.error(`Task ${task.id} failed after 5 attempts:`, error);
      return;
    }

    // Exponential backoff
    const delayMs = Math.min(1000 * Math.pow(2, task.attempts), 300000);
    task.nextAttempt = new Date(Date.now() + delayMs);
  }

  private async executeTask(task: SyncTask): Promise<void> {
    switch (task.type) {
      case 'photo_upload':
        await this.uploadPhoto(task.data as Photo);
        break;
      default:
        throw new Error(`Unknown task type: ${task.type}`);
    }
  }

  private async uploadPhoto(photo: Photo): Promise<void> {
    // Implementation for photo upload
    const uploadResult = await this.apiClient.uploadPhoto(photo);
    await this.photoRepo.markAsSynced(photo.id, uploadResult.url);
  }
}
```

---

## MongoDB Integration Patterns

### PowerSync with TypeScript

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
      new Column({ name: 'sync_status', type: ColumnType.text })
    ]
  })
]);

export class PowerSyncService {
  private db: PowerSyncDatabase;

  constructor() {
    this.db = new PowerSyncDatabase({
      schema,
      database: {
        dbFilename: 'manufacturing.sqlite'
      }
    });
  }

  async uploadPendingPhotos(): Promise<SyncResult[]> {
    const pendingPhotos = await this.db.execute(`
      SELECT * FROM photos 
      WHERE sync_status = 'pending'
      ORDER BY captured_at ASC
      LIMIT 10
    `);

    const results: SyncResult[] = [];

    for (const photo of pendingPhotos.rows) {
      try {
        const uploadResult = await this.uploadPhotoToMongoDB(photo);
        
        await this.db.execute(`
          UPDATE photos 
          SET sync_status = 'synced', remote_url = ?
          WHERE id = ?
        `, [uploadResult.url, photo.id]);

        results.push({ success: true, id: photo.id });
      } catch (error) {
        await this.db.execute(`
          UPDATE photos 
          SET sync_status = 'failed'
          WHERE id = ?
        `, [photo.id]);

        results.push({ success: false, id: photo.id, error });
      }
    }

    return results;
  }

  private async uploadPhotoToMongoDB(photo: any): Promise<{ url: string }> {
    // Implementation for MongoDB upload
    const formData = new FormData();
    formData.append('file', {
      uri: photo.local_path,
      type: 'image/jpeg',
      name: photo.filename
    } as any);

    const response = await fetch('/api/photos/upload', {
      method: 'POST',
      body: formData,
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.statusText}`);
    }

    return await response.json();
  }
}
```

### Type-Safe API Client

```typescript
import { z } from 'zod';

// API response schemas
const ApiResponseSchema = <T extends z.ZodTypeAny>(dataSchema: T) => 
  z.object({
    success: z.boolean(),
    data: dataSchema,
    error: z.string().optional(),
    timestamp: z.string().datetime()
  });

const ProjectSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  jobNumber: z.string(),
  status: z.enum(['active', 'completed']),
  parts: z.array(z.object({
    id: z.string(),
    name: z.string(),
    partNumber: z.string()
  }))
});

export class TypeSafeApiClient {
  constructor(private baseUrl: string) {}

  async getProjects(departmentId: string): Promise<Project[]> {
    const response = await fetch(`${this.baseUrl}/projects?department=${departmentId}`);
    
    if (!response.ok) {
      throw new Error(`API request failed: ${response.statusText}`);
    }

    const rawData = await response.json();
    const validatedResponse = ApiResponseSchema(z.array(ProjectSchema)).parse(rawData);

    if (!validatedResponse.success) {
      throw new Error(validatedResponse.error || 'API request failed');
    }

    return validatedResponse.data;
  }

  async uploadPhoto(photoData: FormData): Promise<{ url: string; id: string }> {
    const response = await fetch(`${this.baseUrl}/photos/upload`, {
      method: 'POST',
      body: photoData
    });

    const rawData = await response.json();
    const uploadResultSchema = z.object({
      url: z.string().url(),
      id: z.string().uuid()
    });

    const validatedResponse = ApiResponseSchema(uploadResultSchema).parse(rawData);

    if (!validatedResponse.success) {
      throw new Error(validatedResponse.error || 'Photo upload failed');
    }

    return validatedResponse.data;
  }
}
```

---

## Performance Optimization Patterns

### Memory Management for Photos

```typescript
export class MemoryEfficientPhotoCache {
  private cache = new Map<string, string>();
  private readonly maxCacheSize = 5;
  private readonly maxPhotoSize = 10 * 1024 * 1024; // 10MB

  async getPhoto(photoId: string): Promise<string | null> {
    // Check cache first
    if (this.cache.has(photoId)) {
      return this.cache.get(photoId)!;
    }

    // Load from storage
    const photoData = await this.loadPhotoFromDisk(photoId);
    
    if (!photoData) return null;

    // Manage cache size
    if (this.cache.size >= this.maxCacheSize) {
      const oldestKey = this.cache.keys().next().value;
      this.cache.delete(oldestKey);
    }

    this.cache.set(photoId, photoData);
    return photoData;
  }

  private async loadPhotoFromDisk(photoId: string): Promise<string | null> {
    try {
      const filePath = `${RNFS.DocumentDirectoryPath}/photos/${photoId}.jpg`;
      const fileExists = await RNFS.exists(filePath);
      
      if (!fileExists) return null;

      const stats = await RNFS.stat(filePath);
      
      if (stats.size > this.maxPhotoSize) {
        console.warn(`Photo ${photoId} is too large: ${stats.size} bytes`);
        return null;
      }

      return await RNFS.readFile(filePath, 'base64');
    } catch (error) {
      console.error(`Failed to load photo ${photoId}:`, error);
      return null;
    }
  }

  clearCache(): void {
    this.cache.clear();
  }
}
```

### Background Processing Queue

```typescript
interface ProcessingJob<T> {
  id: string;
  type: string;
  data: T;
  priority: number;
  attempts: number;
  maxAttempts: number;
}

export class BackgroundProcessor {
  private jobQueue: ProcessingJob<unknown>[] = [];
  private isProcessing = false;
  private processingInterval?: NodeJS.Timeout;

  constructor(private processInterval = 1000) {}

  addJob<T>(job: Omit<ProcessingJob<T>, 'id' | 'attempts'>): void {
    const processingJob: ProcessingJob<T> = {
      ...job,
      id: `job_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`,
      attempts: 0
    };

    this.jobQueue.push(processingJob);
    this.jobQueue.sort((a, b) => b.priority - a.priority);

    this.startProcessing();
  }

  private startProcessing(): void {
    if (this.processingInterval) return;

    this.processingInterval = setInterval(() => {
      if (this.jobQueue.length === 0) {
        this.stopProcessing();
        return;
      }

      if (!this.isProcessing) {
        this.processNextJob();
      }
    }, this.processInterval);
  }

  private async processNextJob(): Promise<void> {
    if (this.jobQueue.length === 0 || this.isProcessing) return;

    this.isProcessing = true;
    const job = this.jobQueue.shift()!;

    try {
      await this.executeJob(job);
    } catch (error) {
      job.attempts++;
      
      if (job.attempts < job.maxAttempts) {
        // Re-queue for retry
        this.jobQueue.unshift(job);
      } else {
        console.error(`Job ${job.id} failed after ${job.maxAttempts} attempts:`, error);
      }
    } finally {
      this.isProcessing = false;
    }
  }

  private async executeJob(job: ProcessingJob<unknown>): Promise<void> {
    switch (job.type) {
      case 'photo_compression':
        await this.processPhotoCompression(job.data);
        break;
      case 'thumbnail_generation':
        await this.processThumbnailGeneration(job.data);
        break;
      case 'photo_upload':
        await this.processPhotoUpload(job.data);
        break;
      default:
        throw new Error(`Unknown job type: ${job.type}`);
    }
  }

  private stopProcessing(): void {
    if (this.processingInterval) {
      clearInterval(this.processingInterval);
      this.processingInterval = undefined;
    }
  }
}
```

---

## Testing Best Practices

### Unit Testing with TypeScript

```typescript
import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { PhotoRepository } from '../services/PhotoRepository';
import { Database } from '@nozbe/watermelondb';

describe('PhotoRepository', () => {
  let photoRepo: PhotoRepository;
  let mockDatabase: jest.Mocked<Database>;

  beforeEach(() => {
    mockDatabase = {
      write: jest.fn(),
      get: jest.fn()
    } as any;

    photoRepo = new PhotoRepository(mockDatabase);
  });

  it('should create photo with correct metadata', async () => {
    const photoData = {
      localPath: '/test/photo.jpg',
      metadata: {
        department: 'Manufacturing',
        project: 'TestProject',
        capturedAt: new Date()
      },
      capturedAt: new Date()
    };

    const mockPhoto = { id: 'photo-123', ...photoData };
    
    mockDatabase.write.mockResolvedValue(mockPhoto as any);

    const result = await photoRepo.create(photoData);

    expect(result).toEqual(mockPhoto);
    expect(mockDatabase.write).toHaveBeenCalledTimes(1);
  });

  it('should find pending sync photos', async () => {
    const mockPhotos = [
      { id: 'photo-1', syncStatus: 'pending' },
      { id: 'photo-2', syncStatus: 'pending' }
    ];

    const mockCollection = {
      query: jest.fn().mockReturnValue({
        fetch: jest.fn().mockResolvedValue(mockPhotos)
      })
    };

    mockDatabase.get.mockReturnValue(mockCollection as any);

    const result = await photoRepo.findPendingSync();

    expect(result).toEqual(mockPhotos);
    expect(mockDatabase.get).toHaveBeenCalledWith('photos');
  });
});
```

### Integration Testing

```typescript
import { render, fireEvent, waitFor, screen } from '@testing-library/react-native';
import { CameraScreen } from '../screens/CameraScreen';
import { PhotoService } from '../services/PhotoService';

jest.mock('../services/PhotoService');

describe('CameraScreen Integration', () => {
  const mockPhotoService = PhotoService as jest.Mocked<typeof PhotoService>;

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should complete photo capture workflow', async () => {
    const mockCapturePhoto = jest.fn().mockResolvedValue({
      id: 'photo-123',
      localPath: '/test/photo.jpg',
      metadata: {}
    });

    mockPhotoService.prototype.capturePhoto = mockCapturePhoto;

    render(<CameraScreen departmentId="manufacturing" />);

    // Simulate photo capture
    const captureButton = screen.getByTestId('capture-button');
    fireEvent.press(captureButton);

    await waitFor(() => {
      expect(mockCapturePhoto).toHaveBeenCalledTimes(1);
    });

    // Verify success state
    expect(screen.getByText('Photo captured successfully')).toBeTruthy();
  });

  it('should handle camera permission denial', async () => {
    mockPhotoService.prototype.requestPermissions = jest.fn().mockResolvedValue(false);

    render(<CameraScreen departmentId="manufacturing" />);

    await waitFor(() => {
      expect(screen.getByText('Camera permission required')).toBeTruthy();
    });
  });
});
```

---

## Security Best Practices

### Input Validation with Zod

```typescript
import { z } from 'zod';

// Environment variable validation
const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'test', 'production']),
  MONGO_CONNECTION_STRING: z.string().url(),
  PHOTO_UPLOAD_ENDPOINT: z.string().url(),
  API_KEY: z.string().min(32),
  DEVICE_ENCRYPTION_KEY: z.string().length(64)
});

export const env = envSchema.parse(process.env);

// Photo metadata validation
const photoMetadataSchema = z.object({
  departmentId: z.string().min(1),
  projectId: z.string().uuid(),
  partId: z.string().min(1),
  capturedAt: z.date(),
  location: z.object({
    latitude: z.number().min(-90).max(90),
    longitude: z.number().min(-180).max(180)
  }).optional(),
  deviceInfo: z.object({
    model: z.string(),
    os: z.string(),
    appVersion: z.string()
  })
});

export function validatePhotoMetadata(data: unknown): PhotoMetadata {
  return photoMetadataSchema.parse(data);
}
```

### Secure Storage

```typescript
import * as Keychain from 'react-native-keychain';
import CryptoJS from 'crypto-js';

export class SecureStorage {
  private readonly encryptionKey: string;

  constructor(encryptionKey: string) {
    this.encryptionKey = encryptionKey;
  }

  async storeSecurely(key: string, value: string): Promise<void> {
    const encrypted = CryptoJS.AES.encrypt(value, this.encryptionKey).toString();
    
    await Keychain.setInternetCredentials(key, key, encrypted);
  }

  async retrieveSecurely(key: string): Promise<string | null> {
    try {
      const credentials = await Keychain.getInternetCredentials(key);
      
      if (!credentials) return null;

      const decrypted = CryptoJS.AES.decrypt(credentials.password, this.encryptionKey);
      return decrypted.toString(CryptoJS.enc.Utf8);
    } catch (error) {
      console.error('Failed to retrieve secure data:', error);
      return null;
    }
  }

  async removeSecurely(key: string): Promise<void> {
    await Keychain.resetInternetCredentials(key);
  }
}
```

---

## Deployment and Build Optimization

### Production Build Configuration

```javascript
// metro.config.js for React Native
const { getDefaultConfig } = require('metro-config');

module.exports = (async () => {
  const config = await getDefaultConfig(__dirname);
  
  // WatermelonDB support
  config.resolver.platforms = ['native', 'android', 'ios'];
  
  // Bundle optimization
  config.transformer.minifierConfig = {
    keep_fnames: true,
    mangle: {
      keep_fnames: true
    }
  };

  // TypeScript support
  config.resolver.sourceExts.push('tsx', 'ts');

  return config;
})();
```

### Android Build Optimization

```gradle
// android/app/build.gradle
android {
    compileSdkVersion 34
    
    defaultConfig {
        minSdkVersion 21
        targetSdkVersion 34
        
        // Enable code shrinking and obfuscation
        minifyEnabled true
        shrinkResources true
        proguardFiles getDefaultProguardFile('proguard-android.txt'), 'proguard-rules.pro'
        
        // Reduce APK size by excluding unused native libraries
        ndk {
            abiFilters "armeabi-v7a", "arm64-v8a", "x86", "x86_64"
        }
    }
    
    buildTypes {
        release {
            // TypeScript optimizations
            bundleInRelease: true
            
            // Photo capture optimizations
            manifestPlaceholders = [
                camera_permission: "android.permission.CAMERA",
                storage_permission: "android.permission.WRITE_EXTERNAL_STORAGE"
            ]
        }
    }
}
```

---

## Summary and Recommendations

### For Manufacturing/Industrial Apps (Alliance Steel Use Case)

**Recommended Stack**:
- **Framework**: React Native with TypeScript
- **Database**: WatermelonDB for offline-first operation
- **Sync**: PowerSync for MongoDB integration
- **Camera**: react-native-vision-camera for manufacturing-grade photos
- **State Management**: TanStack Query + Zustand

### For Standard Business Apps

**Recommended Stack**:
- **Framework**: Expo with TypeScript (fastest development)
- **Database**: Expo SQLite with type-safe queries
- **Sync**: TanStack Query with optimistic updates
- **State Management**: Zustand with persistence

### For Web-First Teams

**Recommended Stack**:
- **Framework**: Capacitor with TypeScript
- **Database**: IndexedDB with Dexie.js
- **Sync**: Custom sync layer with REST APIs
- **State Management**: React Query + Context

This comprehensive guide provides battle-tested patterns for building production-ready TypeScript mobile applications with proper offline support, type safety, and performance optimization.
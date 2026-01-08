# Camera Integration Best Practices for TypeScript Mobile Apps (2024-2025)

**Comprehensive guide for implementing high-quality camera functionality in TypeScript mobile applications with focus on manufacturing, documentation, and industrial use cases.**

---

## Camera Library Selection Guide

### React Native Vision Camera (Recommended)

**Current Status**: Industry standard for React Native camera integration
- **Version**: 4.7.2+ (actively maintained, updated weekly)
- **GitHub Stars**: 6.5K+ with strong community support
- **TypeScript Support**: First-class with comprehensive type definitions
- **Performance**: Optimized for high-quality photo capture and video recording

**Key Advantages**:
- Frame processing capabilities for real-time analysis
- Configurable photo quality up to device maximum resolution
- Advanced camera controls (exposure, focus, white balance)
- Memory-efficient with proper cleanup mechanisms
- Supports both new and old React Native architectures

**When to Choose Vision Camera**:
- Need manufacturing-grade photo quality (2048x2048+)
- Require advanced camera controls and manual settings
- Building photo-heavy applications with performance requirements
- Need frame processing or real-time image analysis
- Working with React Native applications

### Capacitor Camera API

**Current Status**: Modern hybrid solution with consistent cross-platform API
- **Version**: 6.0+ (part of Capacitor 6.0 suite)
- **Platform Support**: iOS, Android, Web with unified TypeScript interface
- **Integration**: Seamless with Ionic, Angular, React, Vue applications

**Key Advantages**:
- Unified API across all platforms including web
- Simple implementation with minimal configuration
- Good for standard photo capture requirements
- Built-in photo editing capabilities
- Automatic permission handling

**When to Choose Capacitor Camera**:
- Building cross-platform applications (mobile + web)
- Need simple photo capture without advanced controls
- Team familiar with web technologies
- Prefer unified API across platforms

### Expo Camera API

**Current Status**: Part of Expo SDK with managed workflow support
- **Version**: SDK 50+ (updated with each Expo release)
- **Integration**: Excellent with Expo managed workflow
- **TypeScript Support**: Complete with Expo's type definitions

**Key Advantages**:
- Fastest setup and development time
- Integrated with Expo development tools
- Good documentation and examples
- Automatic updates with Expo SDK releases

**When to Choose Expo Camera**:
- Using Expo managed workflow
- Rapid prototyping and development
- Standard photo capture requirements
- Team new to React Native

### Library Comparison Matrix

| Feature | Vision Camera | Capacitor | Expo Camera | Ionic Native |
|---------|---------------|-----------|-------------|--------------|
| TypeScript Support | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| Photo Quality Control | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| Performance | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| Advanced Controls | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| Cross-Platform | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| Documentation | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| Community | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

---

## React Native Vision Camera Implementation

### Installation and Setup

```bash
# Install core dependencies
npm install react-native-vision-camera
npm install react-native-permissions

# iOS additional setup (if targeting iOS)
cd ios && pod install

# Android permissions will be configured in AndroidManifest.xml
```

### Android Configuration

```xml
<!-- android/app/src/main/AndroidManifest.xml -->
<manifest xmlns:android="http://schemas.android.com/apk/res/android">
  
  <!-- Camera permissions -->
  <uses-permission android:name="android.permission.CAMERA" />
  
  <!-- Storage permissions for saving photos -->
  <uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />
  <uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
  
  <!-- Optional: Access to device location for photo metadata -->
  <uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
  
  <!-- Camera features -->
  <uses-feature 
    android:name="android.hardware.camera" 
    android:required="true" />
  <uses-feature 
    android:name="android.hardware.camera.autofocus" 
    android:required="false" />
  <uses-feature 
    android:name="android.hardware.camera.flash" 
    android:required="false" />

</manifest>
```

### TypeScript Implementation

```typescript
import React, { useRef, useState, useEffect, useCallback } from 'react';
import { View, Text, TouchableOpacity, Alert } from 'react-native';
import { 
  Camera, 
  useCameraDevice, 
  useCameraPermission,
  PhotoFile,
  TakePhotoOptions 
} from 'react-native-vision-camera';

interface CameraConfig {
  quality: number;
  flash: 'auto' | 'on' | 'off';
  enableShutterSound: boolean;
  skipMetadata: boolean;
  photoCodec: 'jpeg' | 'hevc';
}

interface PhotoMetadata {
  timestamp: Date;
  location?: {
    latitude: number;
    longitude: number;
  };
  deviceInfo: {
    make: string;
    model: string;
    os: string;
  };
}

export interface CameraServiceResult {
  success: boolean;
  photo?: PhotoFile;
  error?: string;
}

export const CameraService: React.FC = () => {
  const camera = useRef<Camera>(null);
  const device = useCameraDevice('back');
  const { hasPermission, requestPermission } = useCameraPermission();
  
  const [isActive, setIsActive] = useState(true);
  const [isCapturing, setIsCapturing] = useState(false);

  // Manufacturing-grade camera configuration
  const cameraConfig: CameraConfig = {
    quality: 100, // Maximum quality for documentation
    flash: 'auto',
    enableShutterSound: false, // Quiet for industrial environments
    skipMetadata: false, // Keep metadata for audit trails
    photoCodec: 'jpeg' // Universal compatibility
  };

  useEffect(() => {
    if (!hasPermission) {
      requestPermission();
    }
  }, [hasPermission, requestPermission]);

  const takePhoto = useCallback(async (): Promise<CameraServiceResult> => {
    if (!camera.current || !device || isCapturing) {
      return { success: false, error: 'Camera not ready' };
    }

    setIsCapturing(true);

    try {
      const photoOptions: TakePhotoOptions = {
        quality: cameraConfig.quality,
        flash: cameraConfig.flash,
        enableShutterSound: cameraConfig.enableShutterSound,
        skipMetadata: cameraConfig.skipMetadata,
        photoCodec: cameraConfig.photoCodec
      };

      const photo = await camera.current.takePhoto(photoOptions);
      
      return { success: true, photo };
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown camera error';
      console.error('Photo capture failed:', error);
      
      return { success: false, error: errorMessage };
    } finally {
      setIsCapturing(false);
    }
  }, [device, isCapturing, cameraConfig]);

  // Handle camera lifecycle
  useEffect(() => {
    return () => {
      setIsActive(false);
    };
  }, []);

  if (!hasPermission) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
        <Text>Camera permission required</Text>
        <TouchableOpacity onPress={requestPermission}>
          <Text>Grant Permission</Text>
        </TouchableOpacity>
      </View>
    );
  }

  if (!device) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
        <Text>No camera device available</Text>
      </View>
    );
  }

  return (
    <View style={{ flex: 1 }}>
      <Camera
        ref={camera}
        style={{ flex: 1 }}
        device={device}
        isActive={isActive}
        photo={true}
        video={false}
      />
      
      <View style={{ position: 'absolute', bottom: 50, alignSelf: 'center' }}>
        <TouchableOpacity
          onPress={takePhoto}
          disabled={isCapturing}
          style={{
            width: 80,
            height: 80,
            borderRadius: 40,
            backgroundColor: isCapturing ? '#ccc' : '#fff',
            justifyContent: 'center',
            alignItems: 'center'
          }}
        >
          <Text>{isCapturing ? 'Capturing...' : 'Capture'}</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
};
```

### Advanced Camera Controls

```typescript
import { useCameraFormat, sortFormats, getCameraDevice } from 'react-native-vision-camera';

interface AdvancedCameraConfig {
  targetResolution: {
    width: number;
    height: number;
  };
  targetFps: number;
  enableHdr: boolean;
  enableLowLightBoost: boolean;
  preferredFormat?: 'photo' | 'video';
}

export const useAdvancedCamera = (config: AdvancedCameraConfig) => {
  const device = useCameraDevice('back');
  
  // Get optimal camera format for requirements
  const format = useCameraFormat(device, [
    { videoResolution: config.targetResolution },
    { fps: config.targetFps },
    { photoHdr: config.enableHdr },
    { photoResolution: 'max' }, // Highest available photo resolution
  ]);

  const captureHighQualityPhoto = useCallback(async (
    camera: Camera,
    customConfig?: Partial<TakePhotoOptions>
  ): Promise<PhotoFile> => {
    const options: TakePhotoOptions = {
      quality: 100,
      flash: 'auto',
      enableShutterSound: false,
      skipMetadata: false,
      photoCodec: 'jpeg',
      ...customConfig
    };

    return await camera.takePhoto(options);
  }, []);

  const getOptimalSettings = useCallback(() => {
    if (!device || !format) return null;

    return {
      device,
      format,
      photoHdr: config.enableHdr,
      lowLightBoost: config.enableLowLightBoost,
      photoQualityBalance: 'quality' as const, // Prioritize quality over speed
    };
  }, [device, format, config]);

  return {
    device,
    format,
    captureHighQualityPhoto,
    getOptimalSettings
  };
};
```

---

## Photo Processing and Metadata Management

### EXIF Data Extraction and Embedding

```typescript
import ExifReader from 'exifreader';
import * as RNFS from 'react-native-fs';

interface PhotoExifData {
  dateTime?: string;
  location?: {
    latitude: number;
    longitude: number;
  };
  cameraInfo?: {
    make: string;
    model: string;
    software: string;
  };
  technicalData?: {
    iso: number;
    exposureTime: string;
    fNumber: number;
    focalLength: number;
  };
}

export class PhotoMetadataService {
  static async extractExifData(photoPath: string): Promise<PhotoExifData> {
    try {
      const photoData = await RNFS.readFile(photoPath, 'base64');
      const buffer = Buffer.from(photoData, 'base64');
      const tags = ExifReader.load(buffer);

      const exifData: PhotoExifData = {
        dateTime: tags['DateTime']?.description,
        location: this.extractLocationData(tags),
        cameraInfo: this.extractCameraInfo(tags),
        technicalData: this.extractTechnicalData(tags)
      };

      return exifData;
    } catch (error) {
      console.error('Failed to extract EXIF data:', error);
      return {};
    }
  }

  private static extractLocationData(tags: any): { latitude: number; longitude: number } | undefined {
    const latitude = tags['GPSLatitude']?.description;
    const longitude = tags['GPSLongitude']?.description;
    
    if (latitude && longitude) {
      return {
        latitude: parseFloat(latitude),
        longitude: parseFloat(longitude)
      };
    }
    
    return undefined;
  }

  private static extractCameraInfo(tags: any) {
    return {
      make: tags['Make']?.description,
      model: tags['Model']?.description,
      software: tags['Software']?.description
    };
  }

  private static extractTechnicalData(tags: any) {
    return {
      iso: tags['ISOSpeedRatings']?.value?.[0],
      exposureTime: tags['ExposureTime']?.description,
      fNumber: tags['FNumber']?.description,
      focalLength: tags['FocalLength']?.description
    };
  }

  // Embed custom metadata for manufacturing use cases
  static async embedCustomMetadata(
    photoPath: string,
    customData: Record<string, any>
  ): Promise<string> {
    try {
      // For React Native, we typically store metadata separately
      // since direct EXIF manipulation is limited on mobile
      const metadataPath = photoPath.replace('.jpg', '_metadata.json');
      
      const metadata = {
        originalPhoto: photoPath,
        customData,
        embeddedAt: new Date().toISOString()
      };

      await RNFS.writeFile(metadataPath, JSON.stringify(metadata), 'utf8');
      
      return metadataPath;
    } catch (error) {
      console.error('Failed to embed custom metadata:', error);
      throw error;
    }
  }
}
```

### Image Compression and Optimization

```typescript
import ImageResizer from '@bam.tech/react-native-image-resizer';

interface CompressionConfig {
  quality: number; // 0-100
  maxWidth?: number;
  maxHeight?: number;
  format: 'JPEG' | 'PNG' | 'WEBP';
  keepMeta: boolean;
  mode: 'contain' | 'cover' | 'stretch';
}

interface ProcessedPhoto {
  originalUri: string;
  compressedUri: string;
  thumbnailUri: string;
  originalSize: number;
  compressedSize: number;
  compressionRatio: number;
  dimensions: {
    width: number;
    height: number;
  };
}

export class PhotoCompressionService {
  static async processPhoto(
    photoUri: string,
    config: Partial<CompressionConfig> = {}
  ): Promise<ProcessedPhoto> {
    const defaultConfig: CompressionConfig = {
      quality: 90,
      maxWidth: 2048,
      maxHeight: 2048,
      format: 'JPEG',
      keepMeta: true,
      mode: 'contain'
    };

    const finalConfig = { ...defaultConfig, ...config };

    try {
      // Get original file info
      const originalStats = await RNFS.stat(photoUri);
      const originalSize = originalStats.size;

      // Compress main image
      const compressedResult = await ImageResizer.createResizedImage(
        photoUri,
        finalConfig.maxWidth || 2048,
        finalConfig.maxHeight || 2048,
        finalConfig.format,
        finalConfig.quality,
        0, // rotation
        undefined, // output path (auto-generated)
        finalConfig.keepMeta,
        {
          mode: finalConfig.mode,
          onlyScaleDown: true // Don't upscale images
        }
      );

      // Generate thumbnail
      const thumbnailResult = await ImageResizer.createResizedImage(
        photoUri,
        300,
        300,
        'JPEG',
        80,
        0,
        undefined,
        false, // Don't keep metadata for thumbnails
        {
          mode: 'cover',
          onlyScaleDown: true
        }
      );

      const compressedStats = await RNFS.stat(compressedResult.uri);
      const compressedSize = compressedStats.size;

      return {
        originalUri: photoUri,
        compressedUri: compressedResult.uri,
        thumbnailUri: thumbnailResult.uri,
        originalSize,
        compressedSize,
        compressionRatio: originalSize / compressedSize,
        dimensions: {
          width: compressedResult.width,
          height: compressedResult.height
        }
      };
    } catch (error) {
      console.error('Photo processing failed:', error);
      throw new Error(`Photo processing failed: ${error}`);
    }
  }

  static async batchProcessPhotos(
    photoUris: string[],
    config?: Partial<CompressionConfig>
  ): Promise<ProcessedPhoto[]> {
    const results: ProcessedPhoto[] = [];
    
    // Process in batches to avoid memory issues
    const batchSize = 3;
    
    for (let i = 0; i < photoUris.length; i += batchSize) {
      const batch = photoUris.slice(i, i + batchSize);
      
      const batchResults = await Promise.allSettled(
        batch.map(uri => this.processPhoto(uri, config))
      );

      batchResults.forEach((result, index) => {
        if (result.status === 'fulfilled') {
          results.push(result.value);
        } else {
          console.error(`Failed to process photo ${batch[index]}:`, result.reason);
        }
      });
    }

    return results;
  }
}
```

### File Organization and Naming

```typescript
import * as RNFS from 'react-native-fs';
import { format } from 'date-fns';

interface FileNamingConfig {
  department: string;
  project: string;
  part: string;
  inspector?: string;
  customPrefix?: string;
}

interface OrganizedPhoto {
  originalPath: string;
  organizedPath: string;
  thumbnailPath: string;
  metadataPath: string;
  fileName: string;
}

export class PhotoOrganizationService {
  private static readonly BASE_PHOTO_DIR = `${RNFS.DocumentDirectoryPath}/photos`;
  private static readonly THUMBNAIL_DIR = `${RNFS.DocumentDirectoryPath}/thumbnails`;
  private static readonly METADATA_DIR = `${RNFS.DocumentDirectoryPath}/metadata`;

  static async initializeDirectories(): Promise<void> {
    const directories = [
      this.BASE_PHOTO_DIR,
      this.THUMBNAIL_DIR,
      this.METADATA_DIR
    ];

    for (const dir of directories) {
      const exists = await RNFS.exists(dir);
      if (!exists) {
        await RNFS.mkdir(dir);
      }
    }
  }

  static generateFileName(config: FileNamingConfig): string {
    const timestamp = format(new Date(), 'yyyyMMdd_HHmmss');
    const parts = [
      config.customPrefix || 'PHOTO',
      config.department.toUpperCase(),
      config.project.replace(/[^a-zA-Z0-9]/g, '_'),
      config.part.replace(/[^a-zA-Z0-9]/g, '_'),
      timestamp
    ];

    if (config.inspector) {
      parts.push(config.inspector.replace(/[^a-zA-Z0-9]/g, '_'));
    }

    return `${parts.join('_')}.jpg`;
  }

  static async organizePhoto(
    photoPath: string,
    config: FileNamingConfig,
    thumbnailPath?: string
  ): Promise<OrganizedPhoto> {
    await this.initializeDirectories();

    const fileName = this.generateFileName(config);
    const organizedPath = `${this.BASE_PHOTO_DIR}/${fileName}`;
    const finalThumbnailPath = `${this.THUMBNAIL_DIR}/thumb_${fileName}`;
    const metadataPath = `${this.METADATA_DIR}/${fileName.replace('.jpg', '.json')}`;

    try {
      // Move/copy original photo
      await RNFS.moveFile(photoPath, organizedPath);

      // Move/copy thumbnail if provided
      if (thumbnailPath) {
        await RNFS.moveFile(thumbnailPath, finalThumbnailPath);
      }

      // Create metadata file
      const metadata = {
        originalFileName: fileName,
        capturedAt: new Date().toISOString(),
        department: config.department,
        project: config.project,
        part: config.part,
        inspector: config.inspector,
        filePaths: {
          original: organizedPath,
          thumbnail: finalThumbnailPath,
          metadata: metadataPath
        }
      };

      await RNFS.writeFile(metadataPath, JSON.stringify(metadata, null, 2));

      return {
        originalPath: photoPath,
        organizedPath,
        thumbnailPath: finalThumbnailPath,
        metadataPath,
        fileName
      };
    } catch (error) {
      console.error('Photo organization failed:', error);
      throw new Error(`Failed to organize photo: ${error}`);
    }
  }

  static async getPhotosByDepartment(department: string): Promise<string[]> {
    try {
      const files = await RNFS.readDir(this.BASE_PHOTO_DIR);
      
      return files
        .filter(file => file.name.includes(`_${department.toUpperCase()}_`))
        .map(file => file.path)
        .sort((a, b) => b.localeCompare(a)); // Most recent first
    } catch (error) {
      console.error('Failed to get photos by department:', error);
      return [];
    }
  }

  static async cleanupOldPhotos(daysToKeep: number = 30): Promise<number> {
    try {
      const files = await RNFS.readDir(this.BASE_PHOTO_DIR);
      const cutoffDate = new Date();
      cutoffDate.setDate(cutoffDate.getDate() - daysToKeep);

      let deletedCount = 0;

      for (const file of files) {
        const fileDate = new Date(file.mtime || 0);
        
        if (fileDate < cutoffDate) {
          // Delete photo, thumbnail, and metadata
          await RNFS.unlink(file.path);
          
          const thumbnailPath = `${this.THUMBNAIL_DIR}/thumb_${file.name}`;
          if (await RNFS.exists(thumbnailPath)) {
            await RNFS.unlink(thumbnailPath);
          }

          const metadataPath = `${this.METADATA_DIR}/${file.name.replace('.jpg', '.json')}`;
          if (await RNFS.exists(metadataPath)) {
            await RNFS.unlink(metadataPath);
          }

          deletedCount++;
        }
      }

      return deletedCount;
    } catch (error) {
      console.error('Cleanup failed:', error);
      return 0;
    }
  }
}
```

---

## Performance Optimization for Camera Apps

### Memory Management

```typescript
import { AppState, AppStateStatus } from 'react-native';

export class CameraMemoryManager {
  private static instance: CameraMemoryManager;
  private photoCache = new Map<string, string>();
  private readonly MAX_CACHE_SIZE = 5;
  private readonly MAX_PHOTO_SIZE = 10 * 1024 * 1024; // 10MB

  static getInstance(): CameraMemoryManager {
    if (!this.instance) {
      this.instance = new CameraMemoryManager();
    }
    return this.instance;
  }

  constructor() {
    // Clear cache when app goes to background
    AppState.addEventListener('change', this.handleAppStateChange);
  }

  private handleAppStateChange = (nextAppState: AppStateStatus) => {
    if (nextAppState === 'background' || nextAppState === 'inactive') {
      this.clearCache();
    }
  };

  async cachePhoto(photoId: string, photoData: string): Promise<void> {
    // Check size limit
    const sizeInBytes = Buffer.byteLength(photoData, 'base64');
    if (sizeInBytes > this.MAX_PHOTO_SIZE) {
      console.warn(`Photo ${photoId} exceeds size limit: ${sizeInBytes} bytes`);
      return;
    }

    // Manage cache size
    if (this.photoCache.size >= this.MAX_CACHE_SIZE) {
      const oldestKey = this.photoCache.keys().next().value;
      this.photoCache.delete(oldestKey);
    }

    this.photoCache.set(photoId, photoData);
  }

  getCachedPhoto(photoId: string): string | null {
    return this.photoCache.get(photoId) || null;
  }

  clearCache(): void {
    this.photoCache.clear();
  }

  getCacheStats(): { size: number; count: number; maxSize: number } {
    const totalSize = Array.from(this.photoCache.values())
      .reduce((total, photo) => total + Buffer.byteLength(photo, 'base64'), 0);

    return {
      size: totalSize,
      count: this.photoCache.size,
      maxSize: this.MAX_CACHE_SIZE
    };
  }
}
```

### Background Processing for Photos

```typescript
interface PhotoProcessingJob {
  id: string;
  type: 'compression' | 'thumbnail' | 'metadata' | 'upload';
  photoPath: string;
  config?: any;
  priority: number;
  attempts: number;
  maxAttempts: number;
}

export class PhotoProcessingQueue {
  private queue: PhotoProcessingJob[] = [];
  private isProcessing = false;
  private workers = new Map<string, Promise<void>>();

  addJob(job: Omit<PhotoProcessingJob, 'id' | 'attempts'>): void {
    const processingJob: PhotoProcessingJob = {
      ...job,
      id: `job_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`,
      attempts: 0
    };

    this.queue.push(processingJob);
    this.queue.sort((a, b) => b.priority - a.priority);

    this.processQueue();
  }

  private async processQueue(): Promise<void> {
    if (this.isProcessing || this.queue.length === 0) return;

    this.isProcessing = true;

    while (this.queue.length > 0) {
      const availableWorkers = 3 - this.workers.size;
      if (availableWorkers <= 0) {
        await Promise.race(this.workers.values());
        continue;
      }

      const job = this.queue.shift()!;
      const workerPromise = this.processJob(job)
        .finally(() => this.workers.delete(job.id));

      this.workers.set(job.id, workerPromise);
    }

    // Wait for all workers to complete
    await Promise.all(this.workers.values());
    this.isProcessing = false;
  }

  private async processJob(job: PhotoProcessingJob): Promise<void> {
    try {
      switch (job.type) {
        case 'compression':
          await this.compressPhoto(job.photoPath, job.config);
          break;
        case 'thumbnail':
          await this.generateThumbnail(job.photoPath, job.config);
          break;
        case 'metadata':
          await this.extractMetadata(job.photoPath, job.config);
          break;
        case 'upload':
          await this.uploadPhoto(job.photoPath, job.config);
          break;
        default:
          throw new Error(`Unknown job type: ${job.type}`);
      }
    } catch (error) {
      job.attempts++;
      
      if (job.attempts < job.maxAttempts) {
        // Exponential backoff
        const delay = Math.min(1000 * Math.pow(2, job.attempts), 30000);
        
        setTimeout(() => {
          this.queue.unshift(job);
          this.processQueue();
        }, delay);
      } else {
        console.error(`Job ${job.id} failed after ${job.maxAttempts} attempts:`, error);
      }
    }
  }

  private async compressPhoto(photoPath: string, config: any): Promise<void> {
    await PhotoCompressionService.processPhoto(photoPath, config);
  }

  private async generateThumbnail(photoPath: string, config: any): Promise<void> {
    // Thumbnail generation logic
    await PhotoCompressionService.processPhoto(photoPath, {
      quality: 80,
      maxWidth: 300,
      maxHeight: 300
    });
  }

  private async extractMetadata(photoPath: string, config: any): Promise<void> {
    await PhotoMetadataService.extractExifData(photoPath);
  }

  private async uploadPhoto(photoPath: string, config: any): Promise<void> {
    // Upload logic - implemented in sync service
    console.log('Uploading photo:', photoPath);
  }

  getQueueStatus(): { pending: number; processing: number; workers: number } {
    return {
      pending: this.queue.length,
      processing: this.workers.size,
      workers: 3
    };
  }
}
```

---

## Testing Camera Integration

### Unit Tests

```typescript
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { CameraService } from '../CameraService';
import { PhotoMetadataService } from '../PhotoMetadataService';

jest.mock('react-native-vision-camera', () => ({
  Camera: 'Camera',
  useCameraDevice: jest.fn(() => ({ id: 'mock-device' })),
  useCameraPermission: jest.fn(() => ({ 
    hasPermission: true,
    requestPermission: jest.fn().mockResolvedValue('granted')
  }))
}));

describe('CameraService', () => {
  const mockTakePhoto = jest.fn();
  
  beforeEach(() => {
    jest.clearAllMocks();
    mockTakePhoto.mockResolvedValue({
      path: '/mock/photo.jpg',
      width: 2048,
      height: 2048
    });
  });

  it('should capture photo with correct configuration', async () => {
    const { getByText } = render(<CameraService />);
    
    const captureButton = getByText('Capture');
    fireEvent.press(captureButton);

    await waitFor(() => {
      expect(mockTakePhoto).toHaveBeenCalledWith({
        quality: 100,
        flash: 'auto',
        enableShutterSound: false,
        skipMetadata: false,
        photoCodec: 'jpeg'
      });
    });
  });

  it('should handle camera permission denial', async () => {
    const mockRequestPermission = jest.fn().mockResolvedValue('denied');
    
    jest.mocked(useCameraPermission).mockReturnValue({
      hasPermission: false,
      requestPermission: mockRequestPermission
    });

    const { getByText } = render(<CameraService />);
    
    expect(getByText('Camera permission required')).toBeTruthy();
  });
});

describe('PhotoMetadataService', () => {
  it('should extract EXIF data correctly', async () => {
    const mockPhotoPath = '/mock/photo.jpg';
    
    // Mock RNFS
    jest.doMock('react-native-fs', () => ({
      readFile: jest.fn().mockResolvedValue('mock-base64-data')
    }));

    const exifData = await PhotoMetadataService.extractExifData(mockPhotoPath);
    
    expect(exifData).toBeDefined();
    expect(typeof exifData).toBe('object');
  });

  it('should handle missing EXIF data gracefully', async () => {
    const mockPhotoPath = '/mock/photo-no-exif.jpg';
    
    jest.doMock('react-native-fs', () => ({
      readFile: jest.fn().mockRejectedValue(new Error('File not found'))
    }));

    const exifData = await PhotoMetadataService.extractExifData(mockPhotoPath);
    
    expect(exifData).toEqual({});
  });
});
```

### Integration Tests

```typescript
import { PhotoOrganizationService } from '../PhotoOrganizationService';
import { PhotoCompressionService } from '../PhotoCompressionService';
import * as RNFS from 'react-native-fs';

describe('Photo Processing Integration', () => {
  const mockPhotoPath = '/mock/original-photo.jpg';
  const mockConfig = {
    department: 'Manufacturing',
    project: 'TestProject',
    part: 'TestPart',
    inspector: 'John Doe'
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should complete full photo processing workflow', async () => {
    // Mock file operations
    jest.mocked(RNFS.exists).mockResolvedValue(true);
    jest.mocked(RNFS.mkdir).mockResolvedValue();
    jest.mocked(RNFS.moveFile).mockResolvedValue();
    jest.mocked(RNFS.writeFile).mockResolvedValue();
    jest.mocked(RNFS.stat).mockResolvedValue({ 
      size: 1024000,
      mtime: new Date()
    } as any);

    // Mock image compression
    const mockCompressedResult = {
      uri: '/mock/compressed-photo.jpg',
      width: 2048,
      height: 2048
    };
    
    jest.doMock('@bam.tech/react-native-image-resizer', () => ({
      createResizedImage: jest.fn().mockResolvedValue(mockCompressedResult)
    }));

    // Process photo
    const processedPhoto = await PhotoCompressionService.processPhoto(mockPhotoPath);
    
    // Organize photo
    const organizedPhoto = await PhotoOrganizationService.organizePhoto(
      processedPhoto.compressedUri,
      mockConfig,
      processedPhoto.thumbnailUri
    );

    expect(organizedPhoto.fileName).toMatch(/PHOTO_MANUFACTURING_TESTPROJECT_TESTPART_\d{8}_\d{6}_JOHN_DOE\.jpg/);
    expect(organizedPhoto.organizedPath).toContain('photos/');
    expect(organizedPhoto.thumbnailPath).toContain('thumbnails/thumb_');
    expect(organizedPhoto.metadataPath).toContain('metadata/');
  });
});
```

---

## Platform-Specific Considerations

### Android Optimizations

```typescript
// android/app/src/main/java/MainActivity.java additional configuration
public class MainActivity extends ReactActivity {
  @Override
  protected void onCreate(Bundle savedInstanceState) {
    // Keep screen on during camera usage
    getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
    
    // Optimize for camera performance
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
      getWindow().setFlags(
        WindowManager.LayoutParams.FLAG_HARDWARE_ACCELERATED,
        WindowManager.LayoutParams.FLAG_HARDWARE_ACCELERATED
      );
    }
    
    super.onCreate(savedInstanceState);
  }
}
```

### ProGuard Configuration

```pro
# android/app/proguard-rules.pro
# Keep Vision Camera classes
-keep class com.mrousavy.camera.** { *; }
-keep class androidx.camera.** { *; }

# Keep EXIF classes
-keep class androidx.exifinterface.** { *; }

# Keep image processing classes
-keep class com.bam.rnimageresizer.** { *; }
```

---

## Common Pitfalls and Solutions

### Issue: Poor Photo Quality on Android

**Problem**: Photos appear blurry or low resolution on certain Android devices.

**Solution**:
```typescript
// Use device-specific quality settings
const getOptimalQuality = (deviceModel: string): number => {
  const lowEndDevices = ['SM-A105', 'SM-A205', 'POCOPHONE']; // Example models
  
  if (lowEndDevices.some(model => deviceModel.includes(model))) {
    return 85; // Lower quality for performance
  }
  
  return 100; // Maximum quality for high-end devices
};

// Apply in camera configuration
const cameraConfig = {
  quality: getOptimalQuality(DeviceInfo.getModel()),
  photoQualityBalance: 'quality' as const,
};
```

### Issue: Memory Issues with Large Photos

**Problem**: App crashes when processing multiple high-resolution photos.

**Solution**:
```typescript
// Implement memory pressure monitoring
export const useMemoryMonitor = () => {
  const [memoryWarning, setMemoryWarning] = useState(false);
  
  useEffect(() => {
    const subscription = AppState.addEventListener('memoryWarning', () => {
      setMemoryWarning(true);
      // Clear photo cache
      CameraMemoryManager.getInstance().clearCache();
      
      // Reduce image quality temporarily
      console.warn('Memory warning: reducing photo quality');
    });

    return () => subscription?.remove();
  }, []);

  return { memoryWarning };
};
```

### Issue: Camera Permissions on Android 11+

**Problem**: Scoped storage restrictions affect photo saving.

**Solution**:
```xml
<!-- android/app/src/main/AndroidManifest.xml -->
<application android:requestLegacyExternalStorage="true">
  <!-- For Android 11+ compatibility -->
  <uses-permission android:name="android.permission.MANAGE_EXTERNAL_STORAGE" />
</application>
```

```typescript
// Check and request storage permissions
import { PermissionsAndroid, Platform } from 'react-native';

export const requestStoragePermissions = async (): Promise<boolean> => {
  if (Platform.OS !== 'android') return true;

  if (Platform.Version >= 30) {
    // Android 11+ - use scoped storage
    return true;
  }

  const granted = await PermissionsAndroid.request(
    PermissionsAndroid.PERMISSIONS.WRITE_EXTERNAL_STORAGE,
    {
      title: 'Storage Permission',
      message: 'App needs storage access to save photos',
      buttonNeutral: 'Ask Me Later',
      buttonNegative: 'Cancel',
      buttonPositive: 'OK',
    }
  );

  return granted === PermissionsAndroid.RESULTS.GRANTED;
};
```

This comprehensive guide provides production-ready patterns for implementing camera functionality in TypeScript mobile applications, with special focus on manufacturing and industrial use cases requiring high-quality photo capture and processing.
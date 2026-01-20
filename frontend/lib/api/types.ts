// Enums
export type FileType = "music" | "photo" | "video" | "document" | "invoice";
export type ProcessingStatus =
  | "pending"
  | "processing"
  | "completed"
  | "failed";
export type SearchType = "fulltext" | "semantic" | "hybrid";

// Tag
export interface Tag {
  id: number;
  user_id?: string;
  name: string;
  color?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTagRequest {
  name: string;
  color?: string;
  description?: string;
}

export interface UpdateTagRequest {
  name?: string;
  color?: string;
  description?: string;
}

// Folder
export interface Folder {
  id: number;
  user_id?: string;
  name: string;
  description?: string;
  parent_id: number | null;
  tags?: Tag[];
  children?: Folder[];
  created_at: string;
  updated_at: string;
}

export interface FolderTree {
  id: number;
  name: string;
  parent_id: number | null;
  children?: FolderTree[];
}

export interface CreateFolderRequest {
  name: string;
  description?: string;
  parent_id?: number | null;
}

export interface UpdateFolderRequest {
  name?: string;
  description?: string;
}

export interface MoveFolderRequest {
  parent_id: number | null;
}

// File
export interface FileItem {
  id: number;
  user_id?: string;
  title: string;
  summary?: string;
  content?: string;
  file_type: FileType;
  folder_id: number | null;
  folder?: Folder;
  tags?: Tag[];
  s3_key: string;
  original_filename: string;
  mime_type?: string;
  size?: number;
  processing_status: ProcessingStatus;
  processing_error?: string;
  has_embedding: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateFileRequest {
  title: string;
  s3_key: string;
  original_filename: string;
  mime_type?: string;
  size?: number;
  folder_id?: number | null;
  file_type?: FileType;
}

export interface UpdateFileRequest {
  title?: string;
  summary?: string;
  file_type?: FileType;
  folder_id?: number | null;
}

export interface MoveFilesRequest {
  file_ids: number[];
  folder_id: number | null;
}

export interface TagIdsRequest {
  tag_ids: number[];
}

// Search
export interface SearchResult {
  file: FileItem;
  score: number;
  snippet?: string;
}

export interface SearchResponse {
  data: SearchResult[];
  total: number;
  query: string;
  search_type: SearchType;
}

export interface SearchOptions {
  q: string;
  type?: SearchType;
  folder_id?: number;
  file_type?: FileType;
  tag_ids?: number[];
  limit?: number;
  offset?: number;
}

// Upload
export interface UploadResponse {
  key: string;
  filename: string;
  size: number;
  content_type: string;
  download_url?: string;
}

export interface PresignedURLResponse {
  upload_url: string;
  key: string;
  content_type: string;
}

export interface FileDownloadURLResponse {
  download_url: string;
  key: string;
  filename: string;
  expires_at: string;
}

// Paginated Response
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

// List Options
export interface FileListOptions {
  folder_id?: number | null;
  file_type?: FileType;
  keyword?: string;
  tag_ids?: number[];
  status?: ProcessingStatus;
  sort_by?: "created_at" | "updated_at" | "title" | "size";
  sort_order?: "asc" | "desc";
  limit?: number;
  offset?: number;
}

export interface FolderListOptions {
  keyword?: string;
  parent_id?: number | null;
  tag_ids?: number[];
  limit?: number;
  offset?: number;
}

export interface TagListOptions {
  keyword?: string;
  limit?: number;
  offset?: number;
}

// API Error
export interface ApiError {
  error: string;
}

// Process Response
export interface ProcessResponse {
  message: string;
  status: ProcessingStatus;
}

// Move Response
export interface MoveResponse {
  message: string;
  moved_count: number;
}

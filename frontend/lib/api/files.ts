import { apiClient } from "./client";
import type {
  FileItem,
  CreateFileRequest,
  UpdateFileRequest,
  PaginatedResponse,
  FileListOptions,
  FileDownloadURLResponse,
  ProcessResponse,
  MoveResponse,
  TagIdsRequest,
} from "./types";

export async function listFiles(
  options: FileListOptions = {},
): Promise<PaginatedResponse<FileItem>> {
  const params = new URLSearchParams();

  if (options.folder_id != null) {
    params.set("folder_id", options.folder_id.toString());
  }
  if (options.file_type) params.set("file_type", options.file_type);
  if (options.keyword) params.set("keyword", options.keyword);
  if (options.tag_ids?.length) {
    params.set("tag_ids", options.tag_ids.join(","));
  }
  if (options.status) params.set("status", options.status);
  if (options.sort_by) params.set("sort_by", options.sort_by);
  if (options.sort_order) params.set("sort_order", options.sort_order);
  if (options.limit) params.set("limit", options.limit.toString());
  if (options.offset) params.set("offset", options.offset.toString());

  const queryString = params.toString();
  const endpoint = `/api/files${queryString ? `?${queryString}` : ""}`;

  return apiClient<PaginatedResponse<FileItem>>(endpoint);
}

export async function getFile(id: number): Promise<FileItem> {
  return apiClient<FileItem>(`/api/files/${id}`);
}

export async function createFile(data: CreateFileRequest): Promise<FileItem> {
  return apiClient<FileItem>("/api/files", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateFile(
  id: number,
  data: UpdateFileRequest,
): Promise<FileItem> {
  return apiClient<FileItem>(`/api/files/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteFile(id: number): Promise<void> {
  return apiClient<void>(`/api/files/${id}`, {
    method: "DELETE",
  });
}

export async function moveFiles(
  fileIds: number[],
  folderId: number | null,
): Promise<MoveResponse> {
  return apiClient<MoveResponse>("/api/files/move", {
    method: "POST",
    body: JSON.stringify({ file_ids: fileIds, folder_id: folderId }),
  });
}

export async function processFile(id: number): Promise<ProcessResponse> {
  return apiClient<ProcessResponse>(`/api/files/${id}/process`, {
    method: "POST",
  });
}

export async function getFileDownloadURL(
  id: number,
): Promise<FileDownloadURLResponse> {
  return apiClient<FileDownloadURLResponse>(`/api/files/${id}/download`);
}

export async function addTagsToFile(
  id: number,
  tagIds: number[],
): Promise<FileItem> {
  const data: TagIdsRequest = { tag_ids: tagIds };
  return apiClient<FileItem>(`/api/files/${id}/tags`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function removeTagsFromFile(
  id: number,
  tagIds: number[],
): Promise<FileItem> {
  const data: TagIdsRequest = { tag_ids: tagIds };
  return apiClient<FileItem>(`/api/files/${id}/tags`, {
    method: "DELETE",
    body: JSON.stringify(data),
  });
}

export async function batchDownloadFiles(fileIds: number[]): Promise<Blob> {
  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL || ""}/api/files/batch-download`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ file_ids: fileIds }),
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Failed to download files");
  }

  return response.blob();
}

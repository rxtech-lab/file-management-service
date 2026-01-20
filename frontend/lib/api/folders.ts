import { apiClient } from "./client";
import type {
  Folder,
  FolderTree,
  CreateFolderRequest,
  UpdateFolderRequest,
  PaginatedResponse,
  FolderListOptions,
  TagIdsRequest,
} from "./types";

export async function listFolders(
  options: FolderListOptions = {},
): Promise<PaginatedResponse<Folder>> {
  const params = new URLSearchParams();

  if (options.keyword) params.set("keyword", options.keyword);
  if (options.parent_id !== undefined && options.parent_id !== null) {
    params.set("parent_id", options.parent_id.toString());
  }
  if (options.tag_ids?.length) {
    params.set("tag_ids", options.tag_ids.join(","));
  }
  if (options.limit) params.set("limit", options.limit.toString());
  if (options.offset) params.set("offset", options.offset.toString());

  const queryString = params.toString();
  const endpoint = `/api/folders${queryString ? `?${queryString}` : ""}`;

  return apiClient<PaginatedResponse<Folder>>(endpoint);
}

export async function getFolderTree(): Promise<FolderTree[]> {
  return apiClient<FolderTree[]>("/api/folders/tree");
}

export async function getFolder(id: number): Promise<Folder> {
  return apiClient<Folder>(`/api/folders/${id}`);
}

export async function createFolder(data: CreateFolderRequest): Promise<Folder> {
  return apiClient<Folder>("/api/folders", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateFolder(
  id: number,
  data: UpdateFolderRequest,
): Promise<Folder> {
  return apiClient<Folder>(`/api/folders/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteFolder(id: number): Promise<void> {
  return apiClient<void>(`/api/folders/${id}`, {
    method: "DELETE",
  });
}

export async function moveFolder(
  id: number,
  parentId: number | null,
): Promise<Folder> {
  return apiClient<Folder>(`/api/folders/${id}/move`, {
    method: "POST",
    body: JSON.stringify({ parent_id: parentId }),
  });
}

export async function addTagsToFolder(
  id: number,
  tagIds: number[],
): Promise<Folder> {
  const data: TagIdsRequest = { tag_ids: tagIds };
  return apiClient<Folder>(`/api/folders/${id}/tags`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function removeTagsFromFolder(
  id: number,
  tagIds: number[],
): Promise<Folder> {
  const data: TagIdsRequest = { tag_ids: tagIds };
  return apiClient<Folder>(`/api/folders/${id}/tags`, {
    method: "DELETE",
    body: JSON.stringify(data),
  });
}

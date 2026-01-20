import { apiClient } from "./client";
import type {
  Tag,
  CreateTagRequest,
  UpdateTagRequest,
  PaginatedResponse,
  TagListOptions,
} from "./types";

export async function listTags(
  options: TagListOptions = {},
): Promise<PaginatedResponse<Tag>> {
  const params = new URLSearchParams();

  if (options.keyword) params.set("keyword", options.keyword);
  if (options.limit) params.set("limit", options.limit.toString());
  if (options.offset) params.set("offset", options.offset.toString());

  const queryString = params.toString();
  const endpoint = `/api/tags${queryString ? `?${queryString}` : ""}`;

  return apiClient<PaginatedResponse<Tag>>(endpoint);
}

export async function getTag(id: number): Promise<Tag> {
  return apiClient<Tag>(`/api/tags/${id}`);
}

export async function createTag(data: CreateTagRequest): Promise<Tag> {
  return apiClient<Tag>("/api/tags", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateTag(
  id: number,
  data: UpdateTagRequest,
): Promise<Tag> {
  return apiClient<Tag>(`/api/tags/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteTag(id: number): Promise<void> {
  return apiClient<void>(`/api/tags/${id}`, {
    method: "DELETE",
  });
}

import { apiClient } from "./client";
import type { SearchResponse, SearchOptions } from "./types";

export async function searchFiles(
  options: SearchOptions,
): Promise<SearchResponse> {
  const params = new URLSearchParams();

  params.set("q", options.q);
  if (options.type) params.set("type", options.type);
  if (options.folder_id) params.set("folder_id", options.folder_id.toString());
  if (options.file_type) params.set("file_type", options.file_type);
  if (options.tag_ids?.length) {
    params.set("tag_ids", options.tag_ids.join(","));
  }
  if (options.limit) params.set("limit", options.limit.toString());
  if (options.offset) params.set("offset", options.offset.toString());

  return apiClient<SearchResponse>(`/api/search?${params.toString()}`);
}

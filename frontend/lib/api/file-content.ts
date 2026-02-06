import { apiClient } from "./client";
import type { FileItem } from "./types";

export interface FileContentResponse {
  content: string;
  mime_type: string;
  size: number;
}

export interface UpdateFileContentRequest {
  content: string;
  mime_type?: string;
}

export async function getFileContent(
  id: number,
): Promise<FileContentResponse> {
  return apiClient<FileContentResponse>(`/api/files/${id}/content`);
}

export async function updateFileContent(
  id: number,
  data: UpdateFileContentRequest,
): Promise<FileItem> {
  return apiClient<FileItem>(`/api/files/${id}/content`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

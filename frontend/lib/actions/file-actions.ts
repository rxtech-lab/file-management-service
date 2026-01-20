"use server";

import { revalidatePath } from "next/cache";
import * as api from "@/lib/api/files";
import type {
  FileItem,
  CreateFileRequest,
  UpdateFileRequest,
  PaginatedResponse,
  FileListOptions,
  FileDownloadURLResponse,
  ProcessResponse,
  MoveResponse,
} from "@/lib/api/types";

interface ActionResult<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export async function listFilesAction(
  options: FileListOptions = {},
): Promise<ActionResult<PaginatedResponse<FileItem>>> {
  try {
    const data = await api.listFiles(options);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to list files",
    };
  }
}

export async function getFileAction(
  id: number,
): Promise<ActionResult<FileItem>> {
  try {
    const data = await api.getFile(id);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to get file",
    };
  }
}

export async function createFileAction(
  data: CreateFileRequest,
): Promise<ActionResult<FileItem>> {
  try {
    const file = await api.createFile(data);
    revalidatePath("/files");
    return { success: true, data: file };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to create file",
    };
  }
}

export async function updateFileAction(
  id: number,
  data: UpdateFileRequest,
): Promise<ActionResult<FileItem>> {
  try {
    const file = await api.updateFile(id, data);
    revalidatePath("/files");
    return { success: true, data: file };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update file",
    };
  }
}

export async function deleteFileAction(
  id: number,
): Promise<ActionResult<void>> {
  try {
    await api.deleteFile(id);
    revalidatePath("/files");
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to delete file",
    };
  }
}

export async function moveFilesAction(
  fileIds: number[],
  folderId: number | null,
): Promise<ActionResult<MoveResponse>> {
  try {
    const result = await api.moveFiles(fileIds, folderId);
    revalidatePath("/files");
    return { success: true, data: result };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to move files",
    };
  }
}

export async function processFileAction(
  id: number,
): Promise<ActionResult<ProcessResponse>> {
  try {
    const result = await api.processFile(id);
    return { success: true, data: result };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to process file",
    };
  }
}

export async function getFileDownloadAction(
  id: number,
): Promise<ActionResult<FileDownloadURLResponse>> {
  try {
    const result = await api.getFileDownloadURL(id);
    return { success: true, data: result };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error ? error.message : "Failed to get download URL",
    };
  }
}

export async function addTagsToFileAction(
  id: number,
  tagIds: number[],
): Promise<ActionResult<FileItem>> {
  try {
    const file = await api.addTagsToFile(id, tagIds);
    revalidatePath("/files");
    return { success: true, data: file };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error ? error.message : "Failed to add tags to file",
    };
  }
}

export async function removeTagsFromFileAction(
  id: number,
  tagIds: number[],
): Promise<ActionResult<FileItem>> {
  try {
    const file = await api.removeTagsFromFile(id, tagIds);
    revalidatePath("/files");
    return { success: true, data: file };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error
          ? error.message
          : "Failed to remove tags from file",
    };
  }
}

export async function batchDownloadFilesAction(
  fileIds: number[],
): Promise<ActionResult<Blob>> {
  try {
    const blob = await api.batchDownloadFiles(fileIds);
    return { success: true, data: blob };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error ? error.message : "Failed to download files",
    };
  }
}

"use server";

import { revalidatePath } from "next/cache";
import * as api from "@/lib/api/folders";
import type {
  Folder,
  FolderTree,
  CreateFolderRequest,
  UpdateFolderRequest,
  PaginatedResponse,
  FolderListOptions,
} from "@/lib/api/types";

interface ActionResult<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export async function listFoldersAction(
  options: FolderListOptions = {},
): Promise<ActionResult<PaginatedResponse<Folder>>> {
  try {
    const data = await api.listFolders(options);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to list folders",
    };
  }
}

export async function getFolderTreeAction(): Promise<
  ActionResult<FolderTree[]>
> {
  try {
    const data = await api.getFolderTree();
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error ? error.message : "Failed to get folder tree",
    };
  }
}

export async function getFolderAction(
  id: number,
): Promise<ActionResult<Folder>> {
  try {
    const data = await api.getFolder(id);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to get folder",
    };
  }
}

export async function createFolderAction(
  data: CreateFolderRequest,
): Promise<ActionResult<Folder>> {
  try {
    const folder = await api.createFolder(data);
    revalidatePath("/files");
    return { success: true, data: folder };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to create folder",
    };
  }
}

export async function updateFolderAction(
  id: number,
  data: UpdateFolderRequest,
): Promise<ActionResult<Folder>> {
  try {
    const folder = await api.updateFolder(id, data);
    revalidatePath("/files");
    return { success: true, data: folder };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update folder",
    };
  }
}

export async function deleteFolderAction(
  id: number,
): Promise<ActionResult<void>> {
  try {
    await api.deleteFolder(id);
    revalidatePath("/files");
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to delete folder",
    };
  }
}

export async function moveFolderAction(
  id: number,
  parentId: number | null,
): Promise<ActionResult<Folder>> {
  try {
    const folder = await api.moveFolder(id, parentId);
    revalidatePath("/files");
    return { success: true, data: folder };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to move folder",
    };
  }
}

export async function addTagsToFolderAction(
  id: number,
  tagIds: number[],
): Promise<ActionResult<Folder>> {
  try {
    const folder = await api.addTagsToFolder(id, tagIds);
    revalidatePath("/files");
    return { success: true, data: folder };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error ? error.message : "Failed to add tags to folder",
    };
  }
}

export async function removeTagsFromFolderAction(
  id: number,
  tagIds: number[],
): Promise<ActionResult<Folder>> {
  try {
    const folder = await api.removeTagsFromFolder(id, tagIds);
    revalidatePath("/files");
    return { success: true, data: folder };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error
          ? error.message
          : "Failed to remove tags from folder",
    };
  }
}

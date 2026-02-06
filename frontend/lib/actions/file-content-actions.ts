"use server";

import { revalidatePath } from "next/cache";
import * as api from "@/lib/api/file-content";
import type { FileItem } from "@/lib/api/types";
import type {
  FileContentResponse,
  UpdateFileContentRequest,
} from "@/lib/api/file-content";

interface ActionResult<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export async function getFileContentAction(
  id: number,
): Promise<ActionResult<FileContentResponse>> {
  try {
    const data = await api.getFileContent(id);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error
          ? error.message
          : "Failed to get file content",
    };
  }
}

export async function updateFileContentAction(
  id: number,
  data: UpdateFileContentRequest,
): Promise<ActionResult<FileItem>> {
  try {
    const file = await api.updateFileContent(id, data);
    revalidatePath(`/preview/${id}`);
    return { success: true, data: file };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error
          ? error.message
          : "Failed to update file content",
    };
  }
}

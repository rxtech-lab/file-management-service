"use server";

import { revalidatePath } from "next/cache";
import * as api from "@/lib/api/tags";
import type {
  Tag,
  CreateTagRequest,
  UpdateTagRequest,
  PaginatedResponse,
  TagListOptions,
} from "@/lib/api/types";

interface ActionResult<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export async function listTagsAction(
  options: TagListOptions = {},
): Promise<ActionResult<PaginatedResponse<Tag>>> {
  try {
    const data = await api.listTags(options);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to list tags",
    };
  }
}

export async function getTagAction(id: number): Promise<ActionResult<Tag>> {
  try {
    const data = await api.getTag(id);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to get tag",
    };
  }
}

export async function createTagAction(
  data: CreateTagRequest,
): Promise<ActionResult<Tag>> {
  try {
    const tag = await api.createTag(data);
    revalidatePath("/files");
    return { success: true, data: tag };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to create tag",
    };
  }
}

export async function updateTagAction(
  id: number,
  data: UpdateTagRequest,
): Promise<ActionResult<Tag>> {
  try {
    const tag = await api.updateTag(id, data);
    revalidatePath("/files");
    return { success: true, data: tag };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update tag",
    };
  }
}

export async function deleteTagAction(id: number): Promise<ActionResult<void>> {
  try {
    await api.deleteTag(id);
    revalidatePath("/files");
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to delete tag",
    };
  }
}

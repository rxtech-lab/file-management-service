"use server";

import * as api from "@/lib/api/search";
import type { SearchResponse, SearchOptions } from "@/lib/api/types";

interface ActionResult<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export async function searchFilesAction(
  options: SearchOptions,
): Promise<ActionResult<SearchResponse>> {
  try {
    const data = await api.searchFiles(options);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to search files",
    };
  }
}

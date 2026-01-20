"use server";

import { triggerOrganizeFile, triggerOrganizeFolder, getAgentStatus } from "@/lib/api/agent";
import type { OrganizeFileResponse, OrganizeFolderResponse, AgentStatusResponse } from "@/lib/api/types";

export async function organizeFileAction(
  fileId: number
): Promise<{ success: boolean; data?: OrganizeFileResponse; error?: string }> {
  try {
    const data = await triggerOrganizeFile(fileId);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to organize file",
    };
  }
}

export async function organizeFolderAction(
  folderId: number
): Promise<{ success: boolean; data?: OrganizeFolderResponse; error?: string }> {
  try {
    const data = await triggerOrganizeFolder(folderId);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to organize folder",
    };
  }
}

export async function getAgentStatusAction(): Promise<{
  success: boolean;
  data?: AgentStatusResponse;
  error?: string;
}> {
  try {
    const data = await getAgentStatus();
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error ? error.message : "Failed to get agent status",
    };
  }
}

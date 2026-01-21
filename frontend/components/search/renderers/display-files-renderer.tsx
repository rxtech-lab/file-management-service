"use client";

import { useRouter } from "next/navigation";
import type { ToolResultRendererProps } from "@rx-lab/dashboard-searching-ui";
import { FileResultCard } from "./file-result-card";
import type { FileType } from "@/lib/api/types";

// Type for the display_files tool output
export interface DisplayFilesOutput {
  files: Array<{
    id: number;
    title: string;
    description?: string;
    file_type: string;
    mime_type?: string;
    folder_id?: number | null;
    folder_name?: string;
  }>;
  summary?: string;
}

/**
 * Component that renders the display_files tool result.
 * Matches the ToolResultRendererProps interface from @rx-lab/dashboard-searching-ui.
 */
export function DisplayFilesRenderer({
  output,
  onAction,
}: ToolResultRendererProps) {
  const router = useRouter();
  const data = output as DisplayFilesOutput;

  const handleFileClick = (file: { id: number; folder_id?: number | null }) => {
    const folderId = file.folder_id;
    const path = folderId
      ? `/files/${folderId}?highlight=${file.id}`
      : `/files?highlight=${file.id}`;

    // Navigate and trigger close action
    router.push(path);
    onAction?.({ type: "close", payload: null });
  };

  if (!data?.files) {
    return null;
  }

  return (
    <div className="space-y-2">
      {data.summary && (
        <p className="text-sm text-muted-foreground px-1">{data.summary}</p>
      )}
      <div className="space-y-2">
        {data.files.map((file) => (
          <FileResultCard
            key={file.id}
            file={{
              id: file.id,
              title: file.title,
              file_type: file.file_type as FileType,
              mime_type: file.mime_type,
              folder_id: file.folder_id ?? null,
              folder: file.folder_name ? { name: file.folder_name } : undefined,
            }}
            description={file.description || ""}
            onClick={() =>
              handleFileClick({ id: file.id, folder_id: file.folder_id })
            }
          />
        ))}
      </div>
    </div>
  );
}

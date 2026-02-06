"use client";

import type { PreviewComponentProps } from "@/lib/plugins/types";
import { MonacoEditor } from "./monaco-editor";

export function YamlPreviewContent({ file, content }: PreviewComponentProps) {
  return (
    <div className="flex flex-col h-full">
      <div className="pb-2">
        <h3 className="text-sm font-medium text-muted-foreground">
          {file.original_filename}
        </h3>
      </div>
      <div className="flex-1 min-h-[400px] border rounded-md overflow-hidden">
        <MonacoEditor value={content} readOnly language="yaml" className="h-full" />
      </div>
    </div>
  );
}

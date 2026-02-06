import { notFound } from "next/navigation";
import { FileWarning } from "lucide-react";
import { getFileAction } from "@/lib/actions/file-actions";
import { getFileContentAction } from "@/lib/actions/file-content-actions";
import { PreviewPageClient } from "./preview-page-client";

const MAX_FILE_SIZE = 4 * 1024 * 1024; // 4MB

interface PreviewPageProps {
  params: Promise<{ fileId: string }>;
  searchParams: Promise<{ plugin?: string }>;
}

export default async function PreviewPage({
  params,
  searchParams,
}: PreviewPageProps) {
  const { fileId } = await params;
  const { plugin: pluginId } = await searchParams;
  const id = parseInt(fileId, 10);

  if (isNaN(id)) notFound();

  const fileResult = await getFileAction(id);
  if (!fileResult.success || !fileResult.data) notFound();

  const file = fileResult.data;

  if (file.size && file.size > MAX_FILE_SIZE) {
    const sizeMB = (file.size / 1024 / 1024).toFixed(1);
    return (
      <div className="flex flex-col items-center justify-center h-full gap-4 text-muted-foreground">
        <div className="flex flex-col items-center gap-3 rounded-lg border bg-background p-8 shadow-sm">
          <FileWarning className="h-10 w-10 text-muted-foreground/70" />
          <p className="text-base font-semibold text-foreground">
            File too large to preview
          </p>
          <p className="text-sm text-muted-foreground">
            This file is {sizeMB} MB. Preview is limited to files under 4 MB.
          </p>
        </div>
      </div>
    );
  }

  const contentResult = await getFileContentAction(id);
  if (!contentResult.success || !contentResult.data) notFound();

  return (
    <PreviewPageClient
      file={file}
      content={contentResult.data.content}
      pluginId={pluginId}
    />
  );
}

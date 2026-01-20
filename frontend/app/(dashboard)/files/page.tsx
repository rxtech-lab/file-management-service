import { FilesPageClient } from "./files-page-client";
import { listFilesAction } from "@/lib/actions/file-actions";
import { listFoldersAction } from "@/lib/actions/folder-actions";

interface FilesPageProps {
  searchParams: Promise<{ highlight?: string }>;
}

export default async function FilesPage({ searchParams }: FilesPageProps) {
  const { highlight } = await searchParams;
  const highlightId = highlight ? parseInt(highlight, 10) : undefined;

  // Fetch data for root folder (no folder_id filter means root level)
  const [filesResult, foldersResult] = await Promise.all([
    listFilesAction({ folder_id: null }),
    listFoldersAction({ parent_id: null }),
  ]);

  const files = filesResult.success ? (filesResult.data?.data ?? []) : [];
  const folders = foldersResult.success ? (foldersResult.data?.data ?? []) : [];

  return (
    <FilesPageClient
      initialFiles={files}
      initialFolders={folders}
      currentFolder={null}
      highlightFileId={highlightId}
    />
  );
}

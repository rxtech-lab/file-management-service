import { FilesPageClient } from "./files-page-client";
import { listFilesAction } from "@/lib/actions/file-actions";
import { listFoldersAction } from "@/lib/actions/folder-actions";

interface FilesPageProps {
  searchParams: Promise<{ highlight?: string; tag_ids?: string }>;
}

export default async function FilesPage({ searchParams }: FilesPageProps) {
  const { highlight, tag_ids } = await searchParams;
  const highlightId = highlight ? parseInt(highlight, 10) : undefined;

  // Parse tag_ids from comma-separated string
  const tagIds = tag_ids
    ? tag_ids
        .split(",")
        .map((id) => parseInt(id, 10))
        .filter((id) => !isNaN(id))
    : undefined;

  // Fetch data - if tag filtering, search all folders and don't show folders
  const [filesResult, foldersResult] = await Promise.all([
    listFilesAction({
      folder_id: tagIds ? undefined : null,
      all_folders: !!tagIds,
      tag_ids: tagIds,
    }),
    // Don't fetch folders when tag filtering - only show matching files
    tagIds ? Promise.resolve({ success: true, data: { data: [] } }) : listFoldersAction({ parent_id: null }),
  ]);

  const files = filesResult.success ? (filesResult.data?.data ?? []) : [];
  const folders = foldersResult.success ? (foldersResult.data?.data ?? []) : [];

  return (
    <FilesPageClient
      initialFiles={files}
      initialFolders={folders}
      currentFolder={null}
      highlightFileId={highlightId}
      activeTagIds={tagIds}
    />
  );
}

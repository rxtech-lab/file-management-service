import { FilesPageClient } from "./files-page-client";
import { listFilesAction } from "@/lib/actions/file-actions";
import { listFoldersAction } from "@/lib/actions/folder-actions";

export default async function FilesPage() {
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
    />
  );
}

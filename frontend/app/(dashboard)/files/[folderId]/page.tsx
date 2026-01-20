import { notFound } from "next/navigation";
import { FilesPageClient } from "../files-page-client";
import { listFilesAction } from "@/lib/actions/file-actions";
import {
  listFoldersAction,
  getFolderAction,
} from "@/lib/actions/folder-actions";
import type { Folder } from "@/lib/api/types";

interface FolderPageProps {
  params: Promise<{ folderId: string }>;
  searchParams: Promise<{ highlight?: string; tag_ids?: string }>;
}

// Helper function to build ancestor chain
async function getAncestors(folder: Folder): Promise<Folder[]> {
  const ancestors: Folder[] = [];
  let currentParentId = folder.parent_id;

  while (currentParentId) {
    const result = await getFolderAction(currentParentId);
    if (result.success && result.data) {
      ancestors.unshift(result.data);
      currentParentId = result.data.parent_id;
    } else {
      break;
    }
  }

  return ancestors;
}

export default async function FolderPage({ params, searchParams }: FolderPageProps) {
  const { folderId } = await params;
  const { highlight, tag_ids } = await searchParams;
  const folderIdNum = parseInt(folderId, 10);
  const highlightId = highlight ? parseInt(highlight, 10) : undefined;

  // Parse tag_ids from comma-separated string
  const tagIds = tag_ids
    ? tag_ids
        .split(",")
        .map((id) => parseInt(id, 10))
        .filter((id) => !isNaN(id))
    : undefined;

  if (isNaN(folderIdNum)) {
    notFound();
  }

  // Fetch all data in parallel where possible
  const [folderResult, filesResult, subfoldersResult] = await Promise.all([
    getFolderAction(folderIdNum),
    listFilesAction({ folder_id: folderIdNum, tag_ids: tagIds }),
    listFoldersAction({ parent_id: folderIdNum }),
  ]);

  if (!folderResult.success || !folderResult.data) {
    notFound();
  }

  const currentFolder = folderResult.data;
  const files = filesResult.success ? (filesResult.data?.data ?? []) : [];
  const folders = subfoldersResult.success
    ? (subfoldersResult.data?.data ?? [])
    : [];

  // Get ancestors for breadcrumb
  const ancestors = await getAncestors(currentFolder);

  return (
    <FilesPageClient
      initialFiles={files}
      initialFolders={folders}
      currentFolder={currentFolder}
      ancestors={ancestors}
      highlightFileId={highlightId}
      activeTagIds={tagIds}
    />
  );
}

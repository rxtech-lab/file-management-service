"use client";

import { useRouter, usePathname } from "next/navigation";
import { FolderOpen, Plus } from "lucide-react";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { FolderTreeItem } from "./folder-tree-item";
import type { FolderTree as FolderTreeType } from "@/lib/api/types";

interface FolderTreeProps {
  folders: FolderTreeType[];
  onNewFolder?: () => void;
  onFolderContextMenu?: (e: React.MouseEvent, folder: FolderTreeType) => void;
}

export function FolderTree({
  folders,
  onNewFolder,
  onFolderContextMenu,
}: FolderTreeProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { state } = useSidebar();

  const isRootActive = pathname === "/files" || pathname === "/files/";
  const isCollapsed = state === "collapsed";

  return (
    <SidebarGroup>
      <SidebarGroupLabel className="flex items-center justify-between">
        <span>Folders</span>
        {onNewFolder && (
          <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5"
            onClick={onNewFolder}
          >
            <Plus className="h-4 w-4" />
          </Button>
        )}
      </SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {/* Root folder / All Files */}
          <SidebarMenuItem>
            <SidebarMenuButton
              isActive={isRootActive}
              onClick={() => router.push("/files")}
            >
              <FolderOpen className="h-4 w-4 shrink-0 text-blue-500" />
              <span>All Files</span>
            </SidebarMenuButton>
          </SidebarMenuItem>

          {/* Folder tree */}
          {folders.map((folder) => (
            <FolderTreeItem
              key={folder.id}
              folder={folder}
              onContextMenu={onFolderContextMenu}
            />
          ))}

          {folders.length === 0 && !isCollapsed && (
            <SidebarMenuItem>
              <div className="px-3 py-2 text-sm text-muted-foreground">
                No folders yet
              </div>
            </SidebarMenuItem>
          )}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}

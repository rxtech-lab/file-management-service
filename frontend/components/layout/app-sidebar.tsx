"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { FolderOpen, Tags } from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarSeparator,
} from "@/components/ui/sidebar";
import { NavUser } from "./nav-user";
import { FolderTree } from "@/components/folders/folder-tree";
import { CreateFolderDialog } from "@/components/folders/create-folder-dialog";
import type { FolderTree as FolderTreeType } from "@/lib/api/types";

interface AppSidebarProps {
  variant?: "sidebar" | "floating" | "inset";
  folderTree?: FolderTreeType[];
}

export function AppSidebar({
  variant = "inset",
  folderTree = [],
}: AppSidebarProps) {
  const pathname = usePathname();
  const [showCreateFolder, setShowCreateFolder] = useState(false);

  const isTagsActive = pathname.startsWith("/tags");

  return (
    <>
      <Sidebar variant={variant} collapsible="icon">
        <SidebarHeader className="border-b">
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild>
                <Link href="/files">
                  <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                    <FolderOpen className="size-4" />
                  </div>
                  <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
                    <span className="truncate font-semibold">File Manager</span>
                    <span className="truncate text-xs text-muted-foreground">
                      Organize your files
                    </span>
                  </div>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>

        <SidebarContent>
          {/* Folder Tree */}
          <FolderTree
            folders={folderTree}
            onNewFolder={() => setShowCreateFolder(true)}
          />

          <SidebarSeparator />

          {/* Tags section */}
          <SidebarGroup>
            <SidebarGroupLabel>Manage</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    asChild
                    isActive={isTagsActive}
                    tooltip="Tags"
                  >
                    <Link href="/tags">
                      <Tags className="size-4" />
                      <span>Tags</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter>
          <NavUser />
        </SidebarFooter>
      </Sidebar>

      <CreateFolderDialog
        open={showCreateFolder}
        onOpenChange={setShowCreateFolder}
      />
    </>
  );
}

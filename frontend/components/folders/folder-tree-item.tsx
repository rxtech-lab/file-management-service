"use client";

import { useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { ChevronRight, Folder as FolderIcon } from "lucide-react";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
} from "@/components/ui/sidebar";
import type { FolderTree } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface FolderTreeItemProps {
  folder: FolderTree;
  level?: number;
  onContextMenu?: (e: React.MouseEvent, folder: FolderTree) => void;
}

export function FolderTreeItem({
  folder,
  level = 0,
  onContextMenu,
}: FolderTreeItemProps) {
  const router = useRouter();
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(false);

  const hasChildren = folder.children && folder.children.length > 0;
  const isActive = pathname === `/files/${folder.id}`;

  const handleClick = () => {
    router.push(`/files/${folder.id}`);
  };

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    if (onContextMenu) {
      onContextMenu(e, folder);
    }
  };

  if (hasChildren) {
    return (
      <Collapsible open={isOpen} onOpenChange={setIsOpen}>
        <SidebarMenuItem>
          <CollapsibleTrigger asChild>
            <SidebarMenuButton
              isActive={isActive}
              onClick={handleClick}
              onContextMenu={handleContextMenu}
              className="w-full justify-start"
              tooltip={folder.name}
            >
              <ChevronRight
                className={cn(
                  "h-4 w-4 shrink-0 transition-transform group-data-[collapsible=icon]:hidden",
                  isOpen && "rotate-90",
                )}
                onClick={(e) => {
                  e.stopPropagation();
                  setIsOpen(!isOpen);
                }}
              />
              <FolderIcon className="h-4 w-4 shrink-0 text-yellow-500" />
              <span className="truncate">{folder.name}</span>
            </SidebarMenuButton>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <SidebarMenuSub>
              {folder.children!.map((child) => (
                <FolderTreeItem
                  key={child.id}
                  folder={child}
                  level={level + 1}
                  onContextMenu={onContextMenu}
                />
              ))}
            </SidebarMenuSub>
          </CollapsibleContent>
        </SidebarMenuItem>
      </Collapsible>
    );
  }

  return (
    <SidebarMenuItem>
      <SidebarMenuButton
        isActive={isActive}
        onClick={handleClick}
        onContextMenu={handleContextMenu}
        tooltip={folder.name}
      >
        <span className="w-4 group-data-[collapsible=icon]:hidden" /> {/* Spacer for alignment */}
        <FolderIcon className="h-4 w-4 shrink-0 text-yellow-500" />
        <span className="truncate">{folder.name}</span>
      </SidebarMenuButton>
    </SidebarMenuItem>
  );
}

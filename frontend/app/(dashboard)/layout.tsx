import { cookies } from "next/headers";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/layout/app-sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";
import { getFolderTreeAction } from "@/lib/actions/folder-actions";
import { listTagsAction } from "@/lib/actions/tag-actions";

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const cookieStore = await cookies();
  const defaultOpen = cookieStore.get("sidebar_state")?.value !== "false";

  // Fetch folder tree and tags for sidebar
  const [folderTreeResult, tagsResult] = await Promise.all([
    getFolderTreeAction(),
    listTagsAction({ limit: 10 }),
  ]);
  const folderTree = folderTreeResult.success
    ? (folderTreeResult.data ?? [])
    : [];
  const tags = tagsResult.success ? (tagsResult.data?.data ?? []) : [];

  return (
    <TooltipProvider>
      <SidebarProvider
        defaultOpen={defaultOpen}
        style={
          {
            "--sidebar-width": "16rem",
          } as React.CSSProperties
        }
      >
        <AppSidebar variant="inset" folderTree={folderTree} tags={tags} />
        <SidebarInset>{children}</SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  );
}

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { SiteHeader } from "@/components/layout/site-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import { TagBadge } from "@/components/tags/tag-badge";
import {
  createTagAction,
  updateTagAction,
  deleteTagAction,
} from "@/lib/actions/tag-actions";
import type { Tag } from "@/lib/api/types";
import { formatDate } from "@/lib/utils";

interface TagsPageClientProps {
  initialTags: Tag[];
}

export function TagsPageClient({ initialTags }: TagsPageClientProps) {
  const router = useRouter();
  const [tags] = useState<Tag[]>(initialTags);
  const [showDialog, setShowDialog] = useState(false);
  const [editingTag, setEditingTag] = useState<Tag | null>(null);
  const [name, setName] = useState("");
  const [color, setColor] = useState("#3b82f6");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const openCreateDialog = () => {
    setEditingTag(null);
    setName("");
    setColor("#3b82f6");
    setDescription("");
    setShowDialog(true);
  };

  const openEditDialog = (tag: Tag) => {
    setEditingTag(tag);
    setName(tag.name);
    setColor(tag.color || "#3b82f6");
    setDescription(tag.description || "");
    setShowDialog(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast.error("Tag name is required");
      return;
    }

    setIsLoading(true);

    try {
      if (editingTag) {
        const result = await updateTagAction(editingTag.id, {
          name: name.trim(),
          color,
          description: description.trim() || undefined,
        });

        if (result.success) {
          toast.success("Tag updated successfully");
          setShowDialog(false);
          router.refresh();
        } else {
          toast.error(result.error || "Failed to update tag");
        }
      } else {
        const result = await createTagAction({
          name: name.trim(),
          color,
          description: description.trim() || undefined,
        });

        if (result.success) {
          toast.success("Tag created successfully");
          setShowDialog(false);
          router.refresh();
        } else {
          toast.error(result.error || "Failed to create tag");
        }
      }
    } catch (error) {
      toast.error("An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async (tag: Tag) => {
    if (
      !confirm(
        `Are you sure you want to delete "${tag.name}"? This action cannot be undone.`,
      )
    ) {
      return;
    }

    try {
      const result = await deleteTagAction(tag.id);
      if (result.success) {
        toast.success("Tag deleted successfully");
        router.refresh();
      } else {
        toast.error(result.error || "Failed to delete tag");
      }
    } catch (error) {
      toast.error("Failed to delete tag");
    }
  };

  return (
    <>
      <SiteHeader />

      <div className="p-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Tags</CardTitle>
            <Button onClick={openCreateDialog}>
              <Plus className="h-4 w-4 mr-2" />
              New Tag
            </Button>
          </CardHeader>
          <CardContent>
            {tags.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <p>No tags yet. Create your first tag to get started.</p>
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Tag</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="w-[100px]">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tags.map((tag) => (
                    <TableRow key={tag.id}>
                      <TableCell>
                        <TagBadge tag={tag} />
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {tag.description || "-"}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {formatDate(new Date(tag.created_at))}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => openEditDialog(tag)}
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleDelete(tag)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>

      <Dialog open={showDialog} onOpenChange={setShowDialog}>
        <DialogContent>
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>
                {editingTag ? "Edit Tag" : "Create New Tag"}
              </DialogTitle>
              <DialogDescription>
                {editingTag
                  ? "Update the tag details below."
                  : "Create a new tag to organize your files."}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Enter tag name"
                  disabled={isLoading}
                  autoFocus
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="color">Color</Label>
                <div className="flex items-center gap-2">
                  <input
                    type="color"
                    id="color"
                    value={color}
                    onChange={(e) => setColor(e.target.value)}
                    className="h-10 w-10 rounded border cursor-pointer"
                    disabled={isLoading}
                  />
                  <Input
                    value={color}
                    onChange={(e) => setColor(e.target.value)}
                    placeholder="#000000"
                    disabled={isLoading}
                    className="flex-1"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description (optional)</Label>
                <Textarea
                  id="description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Enter tag description"
                  disabled={isLoading}
                  rows={3}
                />
              </div>

              {/* Preview */}
              <div className="space-y-2">
                <Label>Preview</Label>
                <TagBadge
                  tag={{
                    id: 0,
                    name: name || "Tag name",
                    color,
                    created_at: "",
                    updated_at: "",
                  }}
                />
              </div>
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowDialog(false)}
                disabled={isLoading}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isLoading}>
                {editingTag ? "Update Tag" : "Create Tag"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}

"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { TagBadge } from "@/components/tags/tag-badge";
import { updateTagAction } from "@/lib/actions/tag-actions";
import { toast } from "sonner";
import type { Tag } from "@/lib/api/types";

interface EditTagDialogProps {
  tag: Tag | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditTagDialog({ tag, open, onOpenChange }: EditTagDialogProps) {
  const router = useRouter();
  const [name, setName] = useState("");
  const [color, setColor] = useState("#3b82f6");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (tag) {
      setName(tag.name);
      setColor(tag.color || "#3b82f6");
      setDescription(tag.description || "");
    }
  }, [tag]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!tag) return;

    if (!name.trim()) {
      toast.error("Tag name is required");
      return;
    }

    setIsLoading(true);

    try {
      const result = await updateTagAction(tag.id, {
        name: name.trim(),
        color,
        description: description.trim() || undefined,
      });

      if (result.success) {
        toast.success("Tag updated successfully");
        onOpenChange(false);
        router.refresh();
      } else {
        toast.error(result.error || "Failed to update tag");
      }
    } catch {
      toast.error("An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Edit Tag</DialogTitle>
            <DialogDescription>Update the tag details below.</DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="edit-tag-name">Name</Label>
              <Input
                id="edit-tag-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Enter tag name"
                disabled={isLoading}
                autoFocus
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-tag-color">Color</Label>
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  id="edit-tag-color"
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
              <Label htmlFor="edit-tag-description">Description (optional)</Label>
              <Textarea
                id="edit-tag-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Enter tag description"
                disabled={isLoading}
                rows={3}
              />
            </div>

            <div className="space-y-2">
              <Label>Preview</Label>
              <TagBadge
                tag={{
                  id: tag?.id || 0,
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
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              Update Tag
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

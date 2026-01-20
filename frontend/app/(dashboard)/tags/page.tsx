import { listTagsAction } from "@/lib/actions/tag-actions";
import { TagsPageClient } from "./tags-page-client";

export default async function TagsPage() {
  const result = await listTagsAction();
  const tags = result.success ? (result.data?.data ?? []) : [];

  return <TagsPageClient initialTags={tags} />;
}

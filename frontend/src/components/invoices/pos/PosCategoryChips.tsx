"use client";

import { Category } from "@/types/category";
import { ProductGroup } from "@/types/product-group";

export function PosCategoryChips({
  categories,
  groups,
  categoryID,
  groupID,
  onCategoryChange,
  onGroupChange,
}: {
  categories: Category[];
  groups: ProductGroup[];
  categoryID: number | null;
  groupID: number | null;
  onCategoryChange: (id: number | null) => void;
  onGroupChange: (id: number | null) => void;
}) {
  return (
    <div className="min-w-0 space-y-2">
      <div className="flex gap-2 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        <Chip
          active={categoryID == null}
          onClick={() => {
            onCategoryChange(null);
            onGroupChange(null);
          }}
          label="Sve"
        />
        {categories.map((category) => (
          <Chip
            key={category.id}
            active={categoryID === category.id}
            onClick={() => {
              onCategoryChange(category.id);
              onGroupChange(null);
            }}
            label={category.name}
          />
        ))}
      </div>

      {categoryID != null && groups.length > 0 ? (
        <div className="flex gap-2 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          <Chip
            active={groupID == null}
            onClick={() => onGroupChange(null)}
            label="Sve grupe"
            subtle
          />
          {groups.map((group) => (
            <Chip
              key={group.id}
              active={groupID === group.id}
              onClick={() => onGroupChange(group.id)}
              label={group.name}
              subtle
            />
          ))}
        </div>
      ) : null}
    </div>
  );
}

function Chip({
  label,
  active,
  onClick,
  subtle = false,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
  subtle?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`shrink-0 rounded-full border px-3 py-1.5 text-xs font-medium transition ${
        active
          ? subtle
            ? "border-[#c4a484] bg-[#f3ebe1] text-stone-900"
            : "border-stone-900 bg-stone-900 text-white"
          : "border-stone-200 bg-white text-stone-600 hover:border-stone-300 hover:bg-stone-50"
      }`}
    >
      {label}
    </button>
  );
}

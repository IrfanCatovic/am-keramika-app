"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { CategoryForm } from "@/components/categories/CategoryForm";
import { CategoryList } from "@/components/categories/CategoryList";
import { ProductGroupForm } from "@/components/categories/ProductGroupForm";
import { ProductGroupList } from "@/components/categories/ProductGroupList";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import {
  createCategory,
  createProductGroup,
  deleteCategory,
  deleteProductGroup,
  fetchCategories,
  fetchProductGroups,
  getApiBusinessMessage,
  updateCategory,
  updateCategoryStatus,
  updateProductGroup,
} from "@/lib/categories-api";
import { Category } from "@/types/category";
import { ProductGroup } from "@/types/product-group";

type CategoryFormState =
  | { open: false }
  | { open: true; mode: "create" | "edit"; category?: Category };

type GroupFormState =
  | { open: false }
  | { open: true; mode: "create" | "edit"; group?: ProductGroup };

type ConfirmState =
  | { open: false }
  | {
      open: true;
      kind: "delete-category" | "delete-group" | "toggle-status";
      category?: Category;
      group?: ProductGroup;
    };

function parseCategoryId(value: string | null): number | null {
  if (!value) {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function resolveSelectedCategory(
  categories: Category[],
  queryCategoryId: number | null,
): Category | null {
  if (categories.length === 0) {
    return null;
  }
  return (
    categories.find((item) => item.id === queryCategoryId) ?? categories[0]
  );
}

export function CategoriesWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const queryCategoryId = parseCategoryId(searchParams.get("categoryID"));

  const [categories, setCategories] = useState<Category[]>([]);
  const [categoriesLoading, setCategoriesLoading] = useState(true);
  const [categoriesError, setCategoriesError] = useState<string | null>(null);

  const [groups, setGroups] = useState<ProductGroup[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [groupsError, setGroupsError] = useState<string | null>(null);
  const [groupsCategoryId, setGroupsCategoryId] = useState<number | null>(null);

  const [busyCategoryId, setBusyCategoryId] = useState<number | null>(null);
  const [busyGroupId, setBusyGroupId] = useState<number | null>(null);

  const [categoryForm, setCategoryForm] = useState<CategoryFormState>({
    open: false,
  });
  const [groupForm, setGroupForm] = useState<GroupFormState>({ open: false });
  const [confirm, setConfirm] = useState<ConfirmState>({ open: false });
  const [formError, setFormError] = useState<string | null>(null);
  const [confirmError, setConfirmError] = useState<string | null>(null);
  const [formLoading, setFormLoading] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);

  const selectedCategory = useMemo(
    () => resolveSelectedCategory(categories, queryCategoryId),
    [categories, queryCategoryId],
  );
  const selectedId = selectedCategory?.id ?? null;

  const syncCategoryQuery = useCallback(
    (categoryID: number | null) => {
      const params = new URLSearchParams(searchParams.toString());
      if (categoryID) {
        params.set("categoryID", String(categoryID));
      } else {
        params.delete("categoryID");
      }
      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, {
        scroll: false,
      });
    },
    [pathname, router, searchParams],
  );

  const applyCategories = useCallback(
    (data: Category[], preferredId?: number | null) => {
      setCategories(data);
      if (data.length === 0) {
        syncCategoryQuery(null);
        return;
      }
      const preferred =
        preferredId && data.some((item) => item.id === preferredId)
          ? preferredId
          : null;
      const next =
        preferred ??
        resolveSelectedCategory(data, queryCategoryId)?.id ??
        data[0].id;
      if (next !== queryCategoryId) {
        syncCategoryQuery(next);
      }
    },
    [queryCategoryId, syncCategoryQuery],
  );

  const loadCategories = useCallback(
    async (preferredId?: number | null) => {
      setCategoriesLoading(true);
      setCategoriesError(null);
      try {
        const data = await fetchCategories(true);
        applyCategories(data, preferredId);
      } catch (error) {
        setCategories([]);
        setCategoriesError(
          getApiBusinessMessage(error, "Nije moguće učitati kategorije."),
        );
      } finally {
        setCategoriesLoading(false);
      }
    },
    [applyCategories],
  );

  const loadGroups = useCallback(async (categoryID: number) => {
    setGroupsLoading(true);
    setGroupsError(null);
    setGroupsCategoryId(categoryID);
    try {
      const data = await fetchProductGroups(categoryID);
      setGroups(data);
    } catch (error) {
      setGroups([]);
      setGroupsError(
        getApiBusinessMessage(error, "Nije moguće učitati grupe proizvoda."),
      );
    } finally {
      setGroupsLoading(false);
    }
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      try {
        const data = await fetchCategories(true);
        if (cancelled) {
          return;
        }
        applyCategories(data);
      } catch (error) {
        if (cancelled) {
          return;
        }
        setCategories([]);
        setCategoriesError(
          getApiBusinessMessage(error, "Nije moguće učitati kategorije."),
        );
      } finally {
        if (!cancelled) {
          setCategoriesLoading(false);
        }
      }
    }

    void run();
    return () => {
      cancelled = true;
    };
  }, [applyCategories]);

  useEffect(() => {
    if (!selectedId) {
      return;
    }

    let cancelled = false;
    const categoryID = selectedId;

    async function run() {
      try {
        const data = await fetchProductGroups(categoryID);
        if (cancelled) {
          return;
        }
        setGroups(data);
        setGroupsCategoryId(categoryID);
        setGroupsError(null);
      } catch (error) {
        if (cancelled) {
          return;
        }
        setGroups([]);
        setGroupsCategoryId(categoryID);
        setGroupsError(
          getApiBusinessMessage(error, "Nije moguće učitati grupe proizvoda."),
        );
      }
    }

    void run();

    return () => {
      cancelled = true;
    };
  }, [selectedId]);

  const visibleGroups =
    selectedId && groupsCategoryId === selectedId ? groups : [];
  const visibleGroupsLoading = Boolean(
    selectedId && (groupsLoading || groupsCategoryId !== selectedId),
  );
  const visibleGroupsError =
    selectedId && groupsCategoryId === selectedId ? groupsError : null;

  function handleSelectCategory(category: Category) {
    syncCategoryQuery(category.id);
  }

  async function handleCategoryFormSubmit(name: string) {
    if (!categoryForm.open) {
      return;
    }
    setFormLoading(true);
    setFormError(null);
    try {
      if (categoryForm.mode === "create") {
        const created = await createCategory(name);
        await loadCategories(created.id);
      } else if (categoryForm.category) {
        await updateCategory(categoryForm.category.id, name);
        await loadCategories(categoryForm.category.id);
      }
      setCategoryForm({ open: false });
    } catch (error) {
      setFormError(
        getApiBusinessMessage(error, "Greška pri čuvanju kategorije."),
      );
    } finally {
      setFormLoading(false);
    }
  }

  async function handleGroupFormSubmit(name: string) {
    if (!groupForm.open || !selectedCategory) {
      return;
    }
    setFormLoading(true);
    setFormError(null);
    try {
      if (groupForm.mode === "create") {
        await createProductGroup(name, selectedCategory.id);
      } else if (groupForm.group) {
        await updateProductGroup(
          groupForm.group.id,
          name,
          selectedCategory.id,
        );
      }
      await loadGroups(selectedCategory.id);
      setGroupForm({ open: false });
    } catch (error) {
      setFormError(getApiBusinessMessage(error, "Greška pri čuvanju grupe."));
    } finally {
      setFormLoading(false);
    }
  }

  async function handleConfirm() {
    if (!confirm.open) {
      return;
    }
    setConfirmLoading(true);
    setConfirmError(null);
    try {
      if (confirm.kind === "delete-category" && confirm.category) {
        setBusyCategoryId(confirm.category.id);
        await deleteCategory(confirm.category.id);
        await loadCategories();
      }

      if (confirm.kind === "toggle-status" && confirm.category) {
        setBusyCategoryId(confirm.category.id);
        await updateCategoryStatus(
          confirm.category.id,
          !confirm.category.isActive,
        );
        await loadCategories(confirm.category.id);
      }

      if (confirm.kind === "delete-group" && confirm.group && selectedCategory) {
        setBusyGroupId(confirm.group.id);
        await deleteProductGroup(confirm.group.id);
        await loadGroups(selectedCategory.id);
      }

      setConfirm({ open: false });
    } catch (error) {
      setConfirmError(
        getApiBusinessMessage(error, "Akcija nije uspjela. Pokušajte ponovo."),
      );
    } finally {
      setConfirmLoading(false);
      setBusyCategoryId(null);
      setBusyGroupId(null);
    }
  }

  const confirmCopy = useMemo(() => {
    if (!confirm.open) {
      return {
        title: "",
        message: "",
        confirmLabel: "Potvrdi",
        tone: "neutral" as const,
      };
    }
    if (confirm.kind === "delete-category" && confirm.category) {
      return {
        title: "Obriši kategoriju",
        message: `Da li ste sigurni da želite obrisati kategoriju „${confirm.category.name}”? Brisanje je moguće samo ako je kategorija potpuno prazna.`,
        confirmLabel: "Obriši",
        tone: "danger" as const,
      };
    }
    if (confirm.kind === "delete-group" && confirm.group) {
      return {
        title: "Obriši grupu",
        message: `Da li ste sigurni da želite obrisati grupu „${confirm.group.name}”? Ako grupa ima proizvode, brisanje neće biti dozvoljeno.`,
        confirmLabel: "Obriši",
        tone: "danger" as const,
      };
    }
    if (confirm.kind === "toggle-status" && confirm.category) {
      const activating = !confirm.category.isActive;
      return {
        title: activating ? "Aktiviraj kategoriju" : "Deaktiviraj kategoriju",
        message: activating
          ? `Aktivirati kategoriju „${confirm.category.name}”?`
          : `Deaktivirati kategoriju „${confirm.category.name}”? Neaktivna kategorija se može pregledati, ali se nove grupe ne mogu dodavati.`,
        confirmLabel: activating ? "Aktiviraj" : "Deaktiviraj",
        tone: "neutral" as const,
      };
    }
    return {
      title: "Potvrda",
      message: "Potvrdite akciju.",
      confirmLabel: "Potvrdi",
      tone: "neutral" as const,
    };
  }, [confirm]);

  const categoryFormKey = categoryForm.open
    ? `${categoryForm.mode}-${categoryForm.category?.id ?? "new"}`
    : "closed";
  const groupFormKey = groupForm.open
    ? `${groupForm.mode}-${groupForm.group?.id ?? "new"}-${selectedId ?? 0}`
    : "closed";

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="dash-enter min-w-0">
        <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
          AM Keramika
        </p>
        <h1 className="mt-1 break-words text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">
          Kategorije i grupe
        </h1>
        <p className="mt-1 max-w-2xl break-words text-sm text-stone-500">
          Organizujte katalog: kategorije lijevo, grupe izabrane kategorije
          desno.
        </p>
      </header>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 lg:gap-5">
        <CategoryList
          categories={categories}
          selectedId={selectedId}
          loading={categoriesLoading}
          error={categoriesError}
          busyId={busyCategoryId}
          onRetry={() => void loadCategories(selectedId)}
          onSelect={handleSelectCategory}
          onCreate={() => {
            setFormError(null);
            setCategoryForm({ open: true, mode: "create" });
          }}
          onEdit={(category) => {
            setFormError(null);
            setCategoryForm({ open: true, mode: "edit", category });
          }}
          onToggleStatus={(category) => {
            setConfirmError(null);
            setConfirm({ open: true, kind: "toggle-status", category });
          }}
          onDelete={(category) => {
            setConfirmError(null);
            setConfirm({ open: true, kind: "delete-category", category });
          }}
        />

        <ProductGroupList
          category={selectedCategory}
          groups={visibleGroups}
          loading={visibleGroupsLoading}
          error={visibleGroupsError}
          busyId={busyGroupId}
          onRetry={() => {
            if (selectedId) {
              void loadGroups(selectedId);
            }
          }}
          onCreate={() => {
            setFormError(null);
            setGroupForm({ open: true, mode: "create" });
          }}
          onEdit={(group) => {
            setFormError(null);
            setGroupForm({ open: true, mode: "edit", group });
          }}
          onDelete={(group) => {
            setConfirmError(null);
            setConfirm({ open: true, kind: "delete-group", group });
          }}
        />
      </div>

      {categoryForm.open ? (
        <CategoryForm
          key={categoryFormKey}
          open
          mode={categoryForm.mode}
          initialName={
            categoryForm.mode === "edit"
              ? (categoryForm.category?.name ?? "")
              : ""
          }
          loading={formLoading}
          error={formError}
          onClose={() => setCategoryForm({ open: false })}
          onSubmit={handleCategoryFormSubmit}
        />
      ) : null}

      {groupForm.open && selectedCategory ? (
        <ProductGroupForm
          key={groupFormKey}
          open
          mode={groupForm.mode}
          categoryName={selectedCategory.name}
          initialName={
            groupForm.mode === "edit" ? (groupForm.group?.name ?? "") : ""
          }
          loading={formLoading}
          error={formError}
          onClose={() => setGroupForm({ open: false })}
          onSubmit={handleGroupFormSubmit}
        />
      ) : null}

      <ConfirmDialog
        open={confirm.open}
        title={confirmCopy.title}
        message={confirmCopy.message}
        confirmLabel={confirmCopy.confirmLabel}
        tone={confirmCopy.tone}
        loading={confirmLoading}
        error={confirmError}
        onClose={() => {
          if (!confirmLoading) {
            setConfirm({ open: false });
          }
        }}
        onConfirm={() => void handleConfirm()}
      />
    </div>
  );
}

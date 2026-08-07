"use client";

import Link from "next/link";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/AuthProvider";
import {
  PricingFieldValues,
  PricingFields,
} from "@/components/products/PricingFields";
import {
  PendingImage,
  ProductImagesField,
} from "@/components/products/ProductImagesField";
import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import {
  fetchCategories,
  fetchProductGroups,
} from "@/lib/categories-api";
import {
  canViewSensitivePricing,
  getDiscountedRawSalePrice,
  getEffectiveSalePrice,
} from "@/lib/product-pricing";
import { formatMoney } from "@/lib/format";
import {
  createProduct,
  deleteProductImage,
  fetchProduct,
  getApiBusinessMessage,
  reorderProductImages,
  setPrimaryProductImage,
  updateProduct,
  uploadProductImages,
} from "@/lib/products-api";
import { Category } from "@/types/category";
import {
  CreateProductPayload,
  Product,
  ProductImage,
  UpdateProductPayload,
} from "@/types/product";
import { ProductGroup } from "@/types/product-group";

function parseOptionalNumber(value: string): number | undefined {
  const trimmed = value.trim().replace(",", ".");
  if (!trimmed) {
    return undefined;
  }
  const parsed = Number(trimmed);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function formatOptionalNumber(value: number | undefined | null): string {
  if (value == null || !Number.isFinite(value)) {
    return "";
  }
  return String(value);
}

type FormState = {
  name: string;
  description: string;
  categoryID: string;
  groupID: string;
  unit: string;
  stockQuantity: string;
  minStockQuantity: string;
  isActive: boolean;
  isOnSale: boolean;
  discountPercent: string;
  showOnHomepage: boolean;
  pricing: PricingFieldValues;
};

const emptyPricing: PricingFieldValues = {
  purchasePrice: "",
  marginPercent: "",
  vatPercent: "",
  salePrice: "",
};

function emptyForm(): FormState {
  return {
    name: "",
    description: "",
    categoryID: "",
    groupID: "",
    unit: "kom",
    stockQuantity: "0",
    minStockQuantity: "0",
    isActive: true,
    isOnSale: false,
    discountPercent: "",
    showOnHomepage: false,
    pricing: { ...emptyPricing },
  };
}

function formFromProduct(product: Product): FormState {
  return {
    name: product.name,
    description: product.description ?? "",
    categoryID: String(product.categoryID),
    groupID: product.groupID ? String(product.groupID) : "",
    unit: product.unit,
    stockQuantity: String(product.stockQuantity),
    minStockQuantity: String(product.minStockQuantity),
    isActive: product.isActive,
    isOnSale: product.isOnSale,
    discountPercent:
      product.discountPercent > 0
        ? formatOptionalNumber(product.discountPercent)
        : "",
    showOnHomepage: product.showOnHomepage,
    pricing: {
      purchasePrice: formatOptionalNumber(product.purchasePrice),
      marginPercent: formatOptionalNumber(product.marginPercent),
      vatPercent: formatOptionalNumber(product.vatPercent),
      salePrice: formatOptionalNumber(product.salePrice),
    },
  };
}

export function ProductForm({
  mode,
  productId,
}: {
  mode: "create" | "edit";
  productId?: number;
}) {
  const router = useRouter();
  const { user } = useAuth();
  const privileged = user ? canViewSensitivePricing(user.role) : false;
  const isWorker = user?.role === "radnik";

  const [form, setForm] = useState<FormState>(emptyForm);
  const [categories, setCategories] = useState<Category[]>([]);
  const [groups, setGroups] = useState<ProductGroup[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [pendingImages, setPendingImages] = useState<PendingImage[]>([]);
  const [existingImages, setExistingImages] = useState<ProductImage[]>([]);
  const [product, setProduct] = useState<Product | null>(null);

  const [bootLoading, setBootLoading] = useState(mode === "edit");
  const [bootError, setBootError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [imagesBusy, setImagesBusy] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [uploadWarning, setUploadWarning] = useState<string | null>(null);

  const selectedCategoryId = Number(form.categoryID) || null;

  const salePreview = useMemo(() => {
    if (!form.isOnSale || !privileged) {
      return null;
    }
    const discount = parseOptionalNumber(form.discountPercent);
    const sale = parseOptionalNumber(form.pricing.salePrice);
    if (discount == null || discount <= 0 || sale == null || sale <= 0) {
      return null;
    }
    return {
      regular: sale,
      discount,
      raw: getDiscountedRawSalePrice(sale, discount),
      effective: getEffectiveSalePrice(sale, true, discount),
    };
  }, [
    form.isOnSale,
    form.discountPercent,
    form.pricing.salePrice,
    privileged,
  ]);

  const categoryOptions = useMemo(() => {
    if (mode === "create") {
      return categories.filter((item) => item.isActive);
    }
    const active = categories.filter((item) => item.isActive);
    const currentId = product?.categoryID;
    if (
      currentId &&
      !active.some((item) => item.id === currentId)
    ) {
      const existing = categories.find((item) => item.id === currentId);
      if (existing) {
        return [existing, ...active];
      }
      if (product?.category) {
        return [
          {
            id: product.category.id,
            name: `${product.category.name} (neaktivna)`,
            slug: product.category.slug,
            isActive: false,
            createdAt: "",
          },
          ...active,
        ];
      }
    }
    return categories;
  }, [categories, mode, product]);

  useEffect(() => {
    let cancelled = false;

    const timer = window.setTimeout(() => {
      async function boot() {
        setBootError(null);
        try {
          const cats = await fetchCategories(true);
          if (cancelled) {
            return;
          }
          setCategories(cats);

          if (mode === "edit" && productId) {
            const loaded = await fetchProduct(productId);
            if (cancelled) {
              return;
            }
            setProduct(loaded);
            setForm(formFromProduct(loaded));
            setExistingImages(loaded.images ?? []);
          }
        } catch (error) {
          if (!cancelled) {
            setBootError(
              getApiBusinessMessage(
                error,
                mode === "edit"
                  ? "Nije moguće učitati proizvod."
                  : "Nije moguće učitati kategorije.",
              ),
            );
          }
        } finally {
          if (!cancelled) {
            setBootLoading(false);
          }
        }
      }
      void boot();
    }, 0);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [mode, productId]);

  useEffect(() => {
    let cancelled = false;

    const timer = window.setTimeout(() => {
      if (!selectedCategoryId) {
        if (!cancelled) {
          setGroups([]);
          setGroupsLoading(false);
        }
        return;
      }

      setGroupsLoading(true);
      void (async () => {
        try {
          const data = await fetchProductGroups(selectedCategoryId);
          if (!cancelled) {
            setGroups(data);
          }
        } catch {
          if (!cancelled) {
            setGroups([]);
          }
        } finally {
          if (!cancelled) {
            setGroupsLoading(false);
          }
        }
      })();
    }, 0);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [selectedCategoryId]);

  const groupSelectOptions = useMemo(() => {
    const list = [...groups];
    if (
      mode === "edit" &&
      product?.group &&
      product.categoryID === selectedCategoryId &&
      !list.some((item) => item.id === product.group!.id)
    ) {
      list.unshift({
        id: product.group.id,
        name: product.group.name,
        slug: product.group.slug,
        categoryID: product.categoryID,
      });
    }
    return list;
  }, [groups, mode, product, selectedCategoryId]);

  function patchForm(patch: Partial<FormState>) {
    setForm((prev) => ({ ...prev, ...patch }));
  }

  function validate(): string | null {
    if (!form.name.trim()) {
      return "Naziv je obavezan.";
    }
    if (!form.categoryID) {
      return "Kategorija je obavezna.";
    }
    if (!form.unit.trim()) {
      return "Jedinica mjere je obavezna.";
    }
    const stock = parseOptionalNumber(form.stockQuantity);
    const minStock = parseOptionalNumber(form.minStockQuantity);
    if (stock == null || stock < 0) {
      return "Stanje na lageru mora biti 0 ili više.";
    }
    if (minStock == null || minStock < 0) {
      return "Minimalno stanje mora biti 0 ili više.";
    }

    if (privileged) {
      const margin = parseOptionalNumber(form.pricing.marginPercent) ?? 0;
      const vat = parseOptionalNumber(form.pricing.vatPercent) ?? 0;
      const purchase = parseOptionalNumber(form.pricing.purchasePrice);
      const sale = parseOptionalNumber(form.pricing.salePrice);
      const calculated = margin > 0 || vat > 0;
      if (calculated) {
        if (purchase == null || purchase <= 0) {
          return "Za automatski obračun nabavna cena mora biti veća od 0.";
        }
      } else if (sale == null || sale <= 0) {
        return "Prodajna cena je obavezna i mora biti veća od 0.";
      }
    } else if (isWorker) {
      if (mode === "create" || product?.pricingMode === "manual") {
        const sale = parseOptionalNumber(form.pricing.salePrice);
        if (sale == null || sale <= 0) {
          return "Prodajna cena je obavezna i mora biti veća od 0.";
        }
      }
    }

    if (form.isOnSale) {
      if (privileged) {
        const discount = parseOptionalNumber(form.discountPercent);
        if (discount == null || discount <= 0) {
          return "Za akciju unesite popust veći od 0%.";
        }
        if (discount >= 100) {
          return "Popust mora biti manji od 100%.";
        }
      } else if ((product?.discountPercent ?? 0) <= 0) {
        return "Akciju može uključiti samo kada je popust već postavljen (developer/šef/menadžer).";
      }
    }

    return null;
  }

  function buildCreatePayload(): CreateProductPayload {
    const payload: CreateProductPayload = {
      name: form.name.trim(),
      categoryID: Number(form.categoryID),
      unit: form.unit.trim(),
      stockQuantity: parseOptionalNumber(form.stockQuantity) ?? 0,
      minStockQuantity: parseOptionalNumber(form.minStockQuantity) ?? 0,
      description: form.description.trim(),
      isOnSale: form.isOnSale,
      showOnHomepage: form.showOnHomepage,
    };
    if (form.groupID) {
      payload.groupID = Number(form.groupID);
    }
    if (privileged) {
      const purchase = parseOptionalNumber(form.pricing.purchasePrice);
      const margin = parseOptionalNumber(form.pricing.marginPercent);
      const vat = parseOptionalNumber(form.pricing.vatPercent);
      const sale = parseOptionalNumber(form.pricing.salePrice);
      const discount = parseOptionalNumber(form.discountPercent);
      if (purchase != null) {
        payload.purchasePrice = purchase;
      }
      if (margin != null) {
        payload.marginPercent = margin;
      }
      if (vat != null) {
        payload.vatPercent = vat;
      }
      if (sale != null) {
        payload.salePrice = sale;
      }
      payload.discountPercent = discount ?? 0;
    } else {
      const sale = parseOptionalNumber(form.pricing.salePrice);
      if (sale != null) {
        payload.salePrice = sale;
      }
    }
    return payload;
  }

  function buildUpdatePayload(): UpdateProductPayload {
    const payload: UpdateProductPayload = {
      name: form.name.trim(),
      categoryID: Number(form.categoryID),
      groupID: form.groupID ? Number(form.groupID) : null,
      unit: form.unit.trim(),
      stockQuantity: parseOptionalNumber(form.stockQuantity) ?? 0,
      minStockQuantity: parseOptionalNumber(form.minStockQuantity) ?? 0,
      description: form.description.trim(),
      isActive: form.isActive,
      isOnSale: form.isOnSale,
      showOnHomepage: form.showOnHomepage,
    };

    if (privileged) {
      payload.purchasePrice =
        parseOptionalNumber(form.pricing.purchasePrice) ?? null;
      payload.marginPercent =
        parseOptionalNumber(form.pricing.marginPercent) ?? 0;
      payload.vatPercent = parseOptionalNumber(form.pricing.vatPercent) ?? 0;
      const sale = parseOptionalNumber(form.pricing.salePrice);
      if (sale != null) {
        payload.salePrice = sale;
      }
      payload.discountPercent = parseOptionalNumber(form.discountPercent) ?? 0;
    } else if (product?.pricingMode !== "calculated") {
      const sale = parseOptionalNumber(form.pricing.salePrice);
      if (sale != null) {
        payload.salePrice = sale;
      }
    }

    return payload;
  }

  async function uploadPending(productId: number): Promise<string | null> {
    if (pendingImages.length === 0) {
      return null;
    }
    try {
      await uploadProductImages(
        productId,
        pendingImages.map((item) => item.file),
      );
      for (const item of pendingImages) {
        URL.revokeObjectURL(item.previewUrl);
      }
      setPendingImages([]);
      return null;
    } catch (error) {
      return getApiBusinessMessage(
        error,
        "Proizvod je sačuvan, ali upload nekih slika nije uspeo.",
      );
    }
  }

  function resetForNext() {
    setForm((prev) => ({
      ...emptyForm(),
      categoryID: prev.categoryID,
      groupID: prev.groupID,
      unit: prev.unit,
      pricing: {
        ...emptyPricing,
        vatPercent: prev.pricing.vatPercent,
      },
    }));
    for (const item of pendingImages) {
      URL.revokeObjectURL(item.previewUrl);
    }
    setPendingImages([]);
    setFormError(null);
  }

  async function submitForm(intent: "save" | "save-and-next") {
    setFormError(null);
    setUploadWarning(null);
    const validationError = validate();
    if (validationError) {
      setFormError(validationError);
      return;
    }

    setSaving(true);
    try {
      if (mode === "create") {
        const created = await createProduct(buildCreatePayload());
        const warning = await uploadPending(created.id);
        if (intent === "save-and-next") {
          resetForNext();
          if (warning) {
            setUploadWarning(warning);
          }
          return;
        }
        if (warning) {
          router.push(
            `/products/${created.id}/edit?uploadWarning=${encodeURIComponent(warning)}`,
          );
          return;
        }
        router.push(`/products/${created.id}/edit`);
        return;
      }

      if (!productId) {
        return;
      }
      const updated = await updateProduct(productId, buildUpdatePayload());
      setProduct(updated);
      setForm(formFromProduct(updated));
      setExistingImages(updated.images ?? []);
      const warning = await uploadPending(productId);
      if (warning) {
        setUploadWarning(warning);
      } else {
        router.push("/products");
      }
    } catch (error) {
      setFormError(
        getApiBusinessMessage(error, "Greška pri čuvanju proizvoda."),
      );
    } finally {
      setSaving(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await submitForm("save");
  }

  async function refreshImages() {
    if (!productId) {
      return;
    }
    const loaded = await fetchProduct(productId);
    setProduct(loaded);
    setExistingImages(loaded.images ?? []);
  }

  async function handleSetPrimary(imageId: number) {
    if (!productId) {
      return;
    }
    setImagesBusy(true);
    setFormError(null);
    try {
      await setPrimaryProductImage(productId, imageId);
      await refreshImages();
    } catch (error) {
      setFormError(
        getApiBusinessMessage(error, "Nije moguće postaviti primarnu sliku."),
      );
    } finally {
      setImagesBusy(false);
    }
  }

  async function handleMove(imageId: number, direction: "up" | "down") {
    if (!productId) {
      return;
    }
    const sorted = [...existingImages].sort(
      (a, b) => a.sortOrder - b.sortOrder || a.id - b.id,
    );
    const index = sorted.findIndex((image) => image.id === imageId);
    if (index < 0) {
      return;
    }
    const target = direction === "up" ? index - 1 : index + 1;
    if (target < 0 || target >= sorted.length) {
      return;
    }
    const next = [...sorted];
    const [item] = next.splice(index, 1);
    next.splice(target, 0, item);
    setImagesBusy(true);
    setFormError(null);
    try {
      const images = await reorderProductImages(
        productId,
        next.map((image) => image.id),
      );
      setExistingImages(images);
    } catch (error) {
      setFormError(
        getApiBusinessMessage(error, "Promena redosleda nije uspela."),
      );
    } finally {
      setImagesBusy(false);
    }
  }

  async function handleDeleteImage(imageId: number) {
    if (!productId) {
      return;
    }
    setImagesBusy(true);
    try {
      await deleteProductImage(productId, imageId);
      await refreshImages();
    } finally {
      setImagesBusy(false);
    }
  }

  async function handleUploadExisting(files: File[]) {
    if (!productId || files.length === 0) {
      return;
    }
    setImagesBusy(true);
    setFormError(null);
    try {
      await uploadProductImages(productId, files);
      await refreshImages();
    } catch (error) {
      setFormError(
        getApiBusinessMessage(error, "Upload slika nije uspeo."),
      );
    } finally {
      setImagesBusy(false);
    }
  }

  if (bootLoading) {
    return (
      <div className="space-y-4">
        <div className="h-16 animate-pulse rounded-2xl bg-stone-100" />
        <ListSkeleton rows={4} />
      </div>
    );
  }

  if (bootError) {
    return (
      <div className="space-y-4">
        <InlineError
          message={bootError}
          onRetry={() => window.location.reload()}
        />
        <Link
          href="/products"
          className="inline-flex min-h-11 items-center text-sm font-medium text-stone-700 underline-offset-2 hover:underline"
        >
          Nazad na listu
        </Link>
      </div>
    );
  }

  if (mode === "edit" && !product) {
    return (
      <EmptyState
        title="Proizvod nije pronađen"
        description="Proverite ID ili se vratite na listu."
        action={
          <Link
            href="/products"
            className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white"
          >
            Nazad
          </Link>
        }
      />
    );
  }

  const workerCalculatedReadonly =
    isWorker &&
    (mode === "edit"
      ? product?.pricingMode === "calculated"
      : false);

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="dash-enter min-w-0">
        <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
          AM Keramika
        </p>
        <h1 className="mt-1 break-words text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">
          {mode === "create" ? "Novi proizvod" : "Uredi proizvod"}
        </h1>
        <p className="mt-1 max-w-2xl break-words text-sm text-stone-500">
          {mode === "create"
            ? "Unesite podatke artikla. Slug se generiše automatski."
            : product
              ? `Slug: ${product.slug}`
              : null}
        </p>
      </header>

      <form
        className="space-y-5 rounded-2xl border border-stone-200/90 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)] sm:p-6"
        onSubmit={(event) => void handleSubmit(event)}
      >
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="sm:col-span-2">
            <label
              htmlFor="product-name"
              className="mb-1.5 block text-sm font-medium text-stone-700"
            >
              Naziv *
            </label>
            <input
              id="product-name"
              value={form.name}
              onChange={(event) => patchForm({ name: event.target.value })}
              disabled={saving}
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
              placeholder="npr. Keramička pločica 30x30"
            />
          </div>

          <div className="sm:col-span-2">
            <label
              htmlFor="product-description"
              className="mb-1.5 block text-sm font-medium text-stone-700"
            >
              Opis
            </label>
            <textarea
              id="product-description"
              rows={3}
              value={form.description}
              onChange={(event) =>
                patchForm({ description: event.target.value })
              }
              disabled={saving}
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
            />
          </div>

          <div>
            <label
              htmlFor="product-category"
              className="mb-1.5 block text-sm font-medium text-stone-700"
            >
              Kategorija *
            </label>
            <select
              id="product-category"
              value={form.categoryID}
              onChange={(event) =>
                patchForm({ categoryID: event.target.value, groupID: "" })
              }
              disabled={saving}
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
            >
              <option value="">Izaberite kategoriju</option>
              {categoryOptions.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                  {!category.isActive ? " (neaktivna)" : ""}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label
              htmlFor="product-group"
              className="mb-1.5 block text-sm font-medium text-stone-700"
            >
              Grupa
            </label>
            <select
              id="product-group"
              value={form.groupID}
              onChange={(event) => patchForm({ groupID: event.target.value })}
              disabled={saving || !form.categoryID || groupsLoading}
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
            >
              <option value="">
                {!form.categoryID
                  ? "Prvo izaberite kategoriju"
                  : groupsLoading
                    ? "Učitavanje..."
                    : "Bez grupe"}
              </option>
              {groupSelectOptions.map((group) => (
                <option key={group.id} value={group.id}>
                  {group.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label
              htmlFor="product-unit"
              className="mb-1.5 block text-sm font-medium text-stone-700"
            >
              Jedinica *
            </label>
            <input
              id="product-unit"
              value={form.unit}
              onChange={(event) => patchForm({ unit: event.target.value })}
              disabled={saving}
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
              placeholder="kom, m2, kut..."
            />
          </div>

          <div>
            <label
              htmlFor="product-stock"
              className="mb-1.5 block text-sm font-medium text-stone-700"
            >
              Stanje *
            </label>
            <input
              id="product-stock"
              type="text"
              inputMode="decimal"
              value={form.stockQuantity}
              onChange={(event) =>
                patchForm({ stockQuantity: event.target.value })
              }
              disabled={saving}
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
            />
          </div>

          <div>
            <label
              htmlFor="product-min-stock"
              className="mb-1.5 block text-sm font-medium text-stone-700"
            >
              Minimalno stanje *
            </label>
            <input
              id="product-min-stock"
              type="text"
              inputMode="decimal"
              value={form.minStockQuantity}
              onChange={(event) =>
                patchForm({ minStockQuantity: event.target.value })
              }
              disabled={saving}
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
            />
          </div>
        </div>

        {privileged ? (
          <PricingFields
            values={form.pricing}
            disabled={saving}
            onChange={(patch) =>
              patchForm({ pricing: { ...form.pricing, ...patch } })
            }
          />
        ) : (
          <section className="space-y-3 rounded-2xl border border-stone-200 bg-stone-50/70 p-4">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="text-sm font-semibold text-stone-900">
                Prodajna cena
              </h3>
              {product ? (
                <span
                  className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${
                    product.pricingMode === "calculated"
                      ? "bg-[#faf6f1] text-[#8a6a45] ring-[#c4a484]/50"
                      : "bg-stone-100 text-stone-600 ring-stone-200"
                  }`}
                >
                  {product.pricingMode === "calculated"
                    ? "Automatski"
                    : "Ručno"}
                </span>
              ) : null}
            </div>
            {workerCalculatedReadonly ? (
              <>
                <p className="text-sm text-stone-600">
                  Cena se automatski obračunava. Promenu nabavne cene,
                  marže ili PDV-a može uraditi menadžer ili šef.
                </p>
                <p className="text-lg font-semibold tabular-nums text-stone-900">
                  {form.pricing.salePrice
                    ? `${form.pricing.salePrice} RSD`
                    : "—"}
                </p>
              </>
            ) : (
              <div>
                <label
                  htmlFor="worker-sale-price"
                  className="mb-1.5 block text-sm font-medium text-stone-700"
                >
                  Prodajna cena *
                </label>
                <input
                  id="worker-sale-price"
                  type="text"
                  inputMode="decimal"
                  value={form.pricing.salePrice}
                  onChange={(event) =>
                    patchForm({
                      pricing: {
                        ...form.pricing,
                        salePrice: event.target.value,
                      },
                    })
                  }
                  disabled={saving}
                  className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
                />
              </div>
            )}
          </section>
        )}

        <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
          {mode === "edit" ? (
            <label className="inline-flex min-h-10 cursor-pointer items-center gap-2 text-sm text-stone-700">
              <input
                type="checkbox"
                checked={form.isActive}
                onChange={(event) =>
                  patchForm({ isActive: event.target.checked })
                }
                disabled={saving}
                className="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-[#c4a484]"
              />
              Aktivan
            </label>
          ) : null}
          {!privileged ? (
            <label className="inline-flex min-h-10 cursor-pointer items-center gap-2 text-sm text-stone-700">
              <input
                type="checkbox"
                checked={form.isOnSale}
                onChange={(event) =>
                  patchForm({ isOnSale: event.target.checked })
                }
                disabled={saving}
                className="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-[#c4a484]"
              />
              Na akciji
            </label>
          ) : null}
          <label className="inline-flex min-h-10 cursor-pointer items-center gap-2 text-sm text-stone-700">
            <input
              type="checkbox"
              checked={form.showOnHomepage}
              onChange={(event) =>
                patchForm({ showOnHomepage: event.target.checked })
              }
              disabled={saving}
              className="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-[#c4a484]"
            />
            Prikaži na početnoj
          </label>
        </div>

        {privileged ? (
          <section className="space-y-3 rounded-2xl border border-stone-200 bg-[#faf8f5] p-4">
            <div>
              <h2 className="text-sm font-semibold text-stone-900">
                Akcija proizvoda
              </h2>
              <p className="mt-1 text-xs text-stone-500">
                Regularna prodajna cena ostaje nepromenjena; kupac plaća
                akcijsku cenu dok je akcija uključena.
              </p>
            </div>
            <label className="inline-flex min-h-10 cursor-pointer items-center gap-2 text-sm text-stone-700">
              <input
                type="checkbox"
                checked={form.isOnSale}
                onChange={(event) =>
                  patchForm({ isOnSale: event.target.checked })
                }
                disabled={saving}
                className="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-[#c4a484]"
              />
              Proizvod je na akciji
            </label>
            {form.isOnSale ? (
              <div className="space-y-3">
                <div>
                  <label
                    htmlFor="discount-percent"
                    className="mb-1.5 block text-sm font-medium text-stone-700"
                  >
                    Popust (%)
                  </label>
                  <input
                    id="discount-percent"
                    type="text"
                    inputMode="decimal"
                    value={form.discountPercent}
                    onChange={(event) =>
                      patchForm({ discountPercent: event.target.value })
                    }
                    disabled={saving}
                    className="w-full max-w-xs rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
                  />
                </div>
                {salePreview ? (
                  <dl className="grid gap-2 text-sm text-stone-700 sm:grid-cols-2">
                    <div>
                      <dt className="text-xs text-stone-500">Regularna cena</dt>
                      <dd className="tabular-nums font-medium text-stone-900">
                        {formatMoney(salePreview.regular)}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs text-stone-500">Popust</dt>
                      <dd className="tabular-nums font-medium text-stone-900">
                        {salePreview.discount}%
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs text-stone-500">
                        Cena nakon popusta
                      </dt>
                      <dd className="tabular-nums font-medium text-stone-900">
                        {formatMoney(salePreview.raw)}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs text-stone-500">
                        Prodajna akcijska cena
                      </dt>
                      <dd className="tabular-nums text-base font-semibold text-stone-900">
                        {formatMoney(salePreview.effective)}
                      </dd>
                    </div>
                  </dl>
                ) : null}
                <p className="text-xs text-stone-500">
                  Akcijska cena se zaokružuje naviše na 10 RSD.
                </p>
              </div>
            ) : null}
          </section>
        ) : null}

        <ProductImagesField
          mode={mode}
          existingImages={existingImages}
          pendingImages={pendingImages}
          onPendingChange={setPendingImages}
          busy={saving || imagesBusy}
          onSetPrimary={mode === "edit" ? handleSetPrimary : undefined}
          onMove={mode === "edit" ? handleMove : undefined}
          onDeleteExisting={mode === "edit" ? handleDeleteImage : undefined}
          onUploadFiles={mode === "edit" ? handleUploadExisting : undefined}
        />

        {formError ? (
          <p className="break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
            {formError}
          </p>
        ) : null}
        {uploadWarning ? (
          <div className="space-y-2 rounded-xl border border-amber-100 bg-amber-50 px-3 py-3 text-sm text-amber-900">
            <p className="break-words">{uploadWarning}</p>
            {mode === "create" && productId == null ? null : (
              <p className="text-xs text-amber-800">
                Proizvod je sačuvan. Možete nastaviti uređivanje slika.
              </p>
            )}
          </div>
        ) : null}

        <div className="flex flex-col-reverse gap-2 border-t border-stone-100 pt-4 sm:flex-row sm:justify-between">
          <Link
            href="/products"
            className="inline-flex min-h-11 items-center justify-center rounded-xl border border-stone-200 px-4 text-sm font-medium text-stone-700 transition hover:bg-stone-50"
          >
            Nazad
          </Link>
          <div className="flex flex-col gap-2 sm:flex-row">
            {mode === "create" ? (
              <button
                type="button"
                disabled={saving}
                onClick={() => void submitForm("save-and-next")}
                className="inline-flex min-h-11 items-center justify-center rounded-xl border border-stone-200 px-4 text-sm font-medium text-stone-800 transition hover:bg-stone-50 disabled:opacity-60"
              >
                {saving ? "Čuvanje..." : "Sačuvaj i dodaj sledeći"}
              </button>
            ) : null}
            <button
              type="submit"
              disabled={saving}
              className="inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white transition hover:bg-stone-800 disabled:opacity-60"
            >
              {saving ? "Čuvanje..." : "Sačuvaj"}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}

function PlaceholderPage({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="space-y-3">
      <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
        {title}
      </h1>
      <p className="text-sm text-slate-500">{description}</p>
      <div className="rounded-xl border border-dashed border-slate-300 bg-white px-5 py-8 text-sm text-slate-500">
        Modul će biti implementiran u narednoj fazi.
      </div>
    </div>
  );
}

export default function ProductsPage() {
  return (
    <PlaceholderPage
      title="Proizvodi"
      description="Upravljanje katalogom proizvoda."
    />
  );
}

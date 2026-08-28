export default function StorefrontLoading() {
  return (
    <div
      role="status"
      aria-label="Učitavanje"
      className="flex min-h-[70vh] items-center justify-center bg-[#f6f4f1]"
    >
      <div className="flex flex-col items-center gap-4 text-stone-500">
        <span
          className="h-10 w-10 animate-spin rounded-full border-2 border-stone-300 border-t-[#8a6a45]"
          aria-hidden
        />
        <span className="text-sm">Učitavanje ponude...</span>
      </div>
    </div>
  );
}

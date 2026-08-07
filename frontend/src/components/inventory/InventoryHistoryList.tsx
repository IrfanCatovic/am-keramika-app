"use client";

import {
  formatSignedQuantity,
  movementTypeLabel,
  signedMovementQuantity,
} from "@/lib/inventory-status";
import { InventoryMovement } from "@/types/inventory";

function movementChangeClass(value: number): string {
  if (value > 0) {
    return "text-emerald-700";
  }
  if (value < 0) {
    return "text-red-700";
  }
  return "text-stone-700";
}

export function InventoryHistoryList({
  movements,
}: {
  movements: InventoryMovement[];
}) {
  if (movements.length === 0) {
    return (
      <div className="rounded-2xl border border-dashed border-stone-300 bg-white px-5 py-10 text-center text-sm text-stone-500">
        Nema zabilježenih kretanja lagera.
      </div>
    );
  }

  return (
    <>
      <div className="hidden overflow-hidden rounded-2xl border border-stone-200 bg-white lg:block">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="border-b border-stone-200 bg-stone-50/80 text-left text-xs uppercase tracking-[0.08em] text-stone-500">
              <th className="px-4 py-3 font-semibold">Datum</th>
              <th className="px-4 py-3 font-semibold">Proizvod</th>
              <th className="px-4 py-3 font-semibold">Promjena</th>
              <th className="px-4 py-3 font-semibold">Tip</th>
              <th className="px-4 py-3 font-semibold">Korisnik</th>
            </tr>
          </thead>
          <tbody>
            {movements.map((movement) => {
              const signed = signedMovementQuantity(
                movement.type,
                movement.quantity,
              );
              return (
                <tr
                  key={movement.id}
                  className="border-b border-stone-100 last:border-b-0"
                >
                  <td className="px-4 py-3 text-stone-600">
                    {movement.createdAt}
                  </td>
                  <td className="px-4 py-3">
                    <p className="font-medium text-stone-900">
                      {movement.productName}
                    </p>
                    {movement.note ? (
                      <p className="mt-0.5 text-xs text-stone-500">
                        {movement.note}
                      </p>
                    ) : null}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`font-semibold tabular-nums ${movementChangeClass(signed)}`}
                    >
                      {formatSignedQuantity(signed)} {movement.productUnit}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-stone-700">
                    {movementTypeLabel(movement.type)}
                  </td>
                  <td className="px-4 py-3 text-stone-600">
                    {movement.createdByUser?.username ?? "—"}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <ul className="space-y-3 lg:hidden">
        {movements.map((movement) => {
          const signed = signedMovementQuantity(
            movement.type,
            movement.quantity,
          );
          return (
            <li
              key={movement.id}
              className="rounded-2xl border border-stone-200 bg-white p-4"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-xs text-stone-500">{movement.createdAt}</p>
                  <p className="mt-1 break-words font-medium text-stone-900">
                    {movement.productName}
                  </p>
                </div>
                <span
                  className={`shrink-0 text-sm font-semibold tabular-nums ${movementChangeClass(signed)}`}
                >
                  {formatSignedQuantity(signed)} {movement.productUnit}
                </span>
              </div>
              <div className="mt-3 flex flex-wrap items-center justify-between gap-2 text-sm">
                <span className="text-stone-600">
                  {movementTypeLabel(movement.type)}
                </span>
                <span className="text-stone-500">
                  {movement.createdByUser?.username ?? "—"}
                </span>
              </div>
              {movement.note ? (
                <p className="mt-2 text-xs text-stone-500">{movement.note}</p>
              ) : null}
            </li>
          );
        })}
      </ul>
    </>
  );
}

'use client';

import { useEffect, useRef, useState } from 'react';

const heroSlides = [
  {
    desktop: '/heroslike/1_pc.webp',
    mobile: '/heroslike/1_mobile.webp',
  },
  {
    desktop: '/heroslike/2_pc.webp',
    mobile: '/heroslike/2_mobile.webp',
  },
  {
    desktop: '/heroslike/3_pc.webp',
    mobile: '/heroslike/3_mobile.webp',
  },
  {
    desktop: '/heroslike/4_pc.webp',
    mobile: '/heroslike/4_mobile.webp',
  },
] as const;

export function StorefrontHeroSlider({ alt }: { alt: string }) {
  const slides = [
    heroSlides[heroSlides.length - 1],
    ...heroSlides,
    heroSlides[0],
  ];
  const [position, setPosition] = useState(1);
  const [transitionEnabled, setTransitionEnabled] = useState(true);
  const touchStartX = useRef<number | null>(null);

  useEffect(() => {
    const interval = window.setInterval(() => {
      setPosition((current) => current + 1);
    }, 5000);

    return () => window.clearInterval(interval);
  }, []);

  const activeSlide = (position - 1 + heroSlides.length) % heroSlides.length;

  function moveTo(positionToShow: number) {
    setTransitionEnabled(true);
    setPosition(positionToShow);
  }

  function handleTransitionEnd() {
    if (position !== 0 && position !== slides.length - 1) return;

    setTransitionEnabled(false);
    setPosition(position === 0 ? heroSlides.length : 1);
    window.requestAnimationFrame(() => {
      window.requestAnimationFrame(() => setTransitionEnabled(true));
    });
  }

  function handleTouchStart(event: React.TouchEvent<HTMLDivElement>) {
    touchStartX.current = event.touches[0]?.clientX ?? null;
  }

  function handleTouchEnd(event: React.TouchEvent<HTMLDivElement>) {
    const startX = touchStartX.current;
    const endX = event.changedTouches[0]?.clientX;
    touchStartX.current = null;
    if (startX == null || endX == null) return;

    const distance = startX - endX;
    if (Math.abs(distance) < 48) return;
    if (distance > 0) {
      moveTo(position + 1);
    } else {
      moveTo(position - 1);
    }
  }

  return (
    <div
      className="absolute inset-0 overflow-hidden touch-pan-y"
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      <div
        className="flex h-full"
        style={{
          width: `${slides.length * 100}%`,
          transform: `translate3d(-${position * (100 / slides.length)}%, 0, 0)`,
          transition: transitionEnabled ? 'transform 700ms ease-out' : 'none',
        }}
        onTransitionEnd={handleTransitionEnd}
      >
        {slides.map((slide, index) => (
          <div
            key={`${slide.desktop}-${index}`}
            className="relative h-full shrink-0"
            style={{ width: `${100 / slides.length}%` }}
          >
            <picture className="absolute inset-0 block">
              <source media="(max-width: 639px)" srcSet={slide.mobile} />
              <img
                src={slide.desktop}
                alt={index === position ? alt : ''}
                className="absolute inset-0 h-full w-full object-cover object-[center_34%] sm:object-[center_40%]"
                loading={index === 1 ? 'eager' : 'lazy'}
                fetchPriority={index === 1 ? 'high' : undefined}
              />
            </picture>
          </div>
        ))}
      </div>

      <div className="absolute inset-x-0 bottom-0 z-10 flex justify-center gap-1.5 pb-3 sm:pb-4">
        {heroSlides.map((_, index) => {
          const isActive = index === activeSlide;
          return (
            <button
              key={index}
              type="button"
              aria-label={`Prikaži hero sliku ${index + 1}`}
              aria-current={isActive ? 'true' : undefined}
              onClick={() => moveTo(index + 1)}
              className="flex h-8 w-8 items-center justify-center rounded-full outline-none focus-visible:ring-2 focus-visible:ring-white/80"
            >
              <span
                key={`${index}-${isActive ? 'active' : 'idle'}`}
                className={`block rounded-full ${
                  isActive
                    ? 'storefront-hero-progress h-2 w-10 bg-[#d7b896]'
                    : 'h-2 w-2 bg-[#d7b896]/75'
                }`}
              />
            </button>
          );
        })}
      </div>
    </div>
  );
}

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

const FIRST_SLIDE_POSITION = 1;
const LAST_SLIDE_POSITION = heroSlides.length + 1;
const SLIDE_DURATION_MS = 5000;

export function StorefrontHeroSlider({ alt }: { alt: string }) {
  const slides = [
    heroSlides[heroSlides.length - 1],
    ...heroSlides,
    heroSlides[0],
  ];
  const [position, setPosition] = useState(FIRST_SLIDE_POSITION);
  const [transitionEnabled, setTransitionEnabled] = useState(true);
  const [visibilityRevision, setVisibilityRevision] = useState(0);
  const touchStartX = useRef<number | null>(null);

  useEffect(() => {
    if (document.visibilityState !== 'visible') return;

    const timeout = window.setTimeout(() => {
      if (position >= LAST_SLIDE_POSITION) {
        setTransitionEnabled(false);
        setPosition(FIRST_SLIDE_POSITION);
        window.requestAnimationFrame(() => {
          window.requestAnimationFrame(() => setTransitionEnabled(true));
        });
        return;
      }

      setPosition(position + 1);
    }, SLIDE_DURATION_MS);

    return () => window.clearTimeout(timeout);
  }, [position, visibilityRevision]);

  useEffect(() => {
    function handleVisibilityChange() {
      if (document.visibilityState !== 'visible') return;

      // A hidden tab can pause CSS transitions and throttle timers. Normalize
      // the loop before allowing the next automatic transition.
      setTransitionEnabled(false);
      setPosition((current) =>
        current <= 0 || current >= LAST_SLIDE_POSITION
          ? FIRST_SLIDE_POSITION
          : current
      );
      setVisibilityRevision((current) => current + 1);
      window.requestAnimationFrame(() => {
        window.requestAnimationFrame(() => setTransitionEnabled(true));
      });
    }

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () =>
      document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, []);

  const activeSlide = (position - 1 + heroSlides.length) % heroSlides.length;

  function moveTo(positionToShow: number) {
    setTransitionEnabled(true);
    setPosition(Math.max(0, Math.min(LAST_SLIDE_POSITION, positionToShow)));
  }

  function handleTransitionEnd() {
    if (position !== 0 && position !== LAST_SLIDE_POSITION) return;

    setTransitionEnabled(false);
    setPosition(position === 0 ? heroSlides.length : FIRST_SLIDE_POSITION);
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
                loading="eager"
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

import { useState, useRef, useCallback } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Container } from "@/components/landing/container";

const REVIEWS = [
  {
    text: "Setting up our workspace took under two minutes. Having a verified sending domain from day one made a huge difference to our deliverability.",
    stars: 5,
    name: "Ryan Almeida",
    role: "Founder, Papertrail",
    initials: "RA",
    color: "#3b7ef5",
    time: "1 day ago",
  },
  {
    text: "We moved off a shared newsletter platform after one too many deliverability issues. The isolation here is real — our own domain, our own dedicated email delivery service.",
    stars: 5,
    name: "Blossom Menezes",
    role: "Head of Growth, Capsule",
    initials: "BM",
    color: "#3f8f6c",
    time: "3 days ago",
  },
  {
    text: "The SQL segmentation alone is worth it. Being able to query subscribers with custom expressions is something we couldn't do on any other hosted platform.",
    stars: 5,
    name: "Jason Park",
    role: "Newsletter operator",
    initials: "JP",
    color: "#d97d3d",
    time: "1 week ago",
  },
  {
    text: "Clean provisioning, dedicated infrastructure, no shared pools. Exactly what we needed to run newsletters for our agency clients without cross-contamination.",
    stars: 5,
    name: "Amara Nwosu",
    role: "Agency owner",
    initials: "AN",
    color: "#8b5cf6",
    time: "2 weeks ago",
  },
  {
    text: "Finally a newsletter tool that takes ownership seriously. Our own domain, our own data, OIDC auth with granular roles. It's everything we needed.",
    stars: 5,
    name: "Theo Lindqvist",
    role: "Engineering lead, Strata",
    initials: "TL",
    color: "#c1503b",
    time: "3 weeks ago",
  },
];

function Stars({ count }: { count: number }) {
  return (
    <div className="flex gap-0.5">
      {Array.from({ length: count }).map((_, i) => (
        <svg key={i} viewBox="0 0 20 20" className="size-4 fill-[#00b67a]">
          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
        </svg>
      ))}
    </div>
  );
}

function ReviewCard({ review }: { review: (typeof REVIEWS)[0] }) {
  return (
    <div
      className="rounded-2xl p-1.5"
      style={{ backgroundColor: `${review.color}22` }}
    >
      <div className="flex flex-col gap-4 rounded-xl border border-white/5 bg-card p-5">
        <p className="text-sm leading-relaxed text-foreground/80 line-clamp-4">
          {review.text}
        </p>
        <div className="flex items-center gap-3">
          <div
            className="flex size-8 shrink-0 items-center justify-center rounded-full text-[11px] font-semibold text-white"
            style={{ backgroundColor: review.color }}
          >
            {review.initials}
          </div>
          <div>
            <p className="text-sm font-medium text-foreground">{review.name}</p>
            <p className="text-xs text-muted-foreground">{review.role}</p>
          </div>
        </div>
      </div>
      <div className="flex justify-center py-2">
        <Stars count={review.stars} />
      </div>
    </div>
  );
}

const DEPTH_TRANSFORMS = [
  "",
  "translateY(12px) rotate(-2.5deg) scale(0.97)",
  "translateY(22px) rotate(4deg) scale(0.94)",
];

export function Reviews() {
  const N = REVIEWS.length;
  const [current, setCurrent] = useState(0);
  const [dragX, setDragX] = useState(0);
  const [dragging, setDragging] = useState(false);
  const [isFlying, setIsFlying] = useState(false);
  const [flyDir, setFlyDir] = useState<1 | -1>(1);
  const startX = useRef(0);
  const flyTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  const dismiss = useCallback(
    (dir: 1 | -1) => {
      if (flyTimeout.current) clearTimeout(flyTimeout.current);
      setFlyDir(dir);
      setIsFlying(true);
      flyTimeout.current = setTimeout(() => {
        setCurrent((c) => (c + 1) % N);
        setIsFlying(false);
        setDragX(0);
      }, 320);
    },
    [N],
  );

  const prev = useCallback(() => {
    if (isFlying) return;
    setCurrent((c) => (c - 1 + N) % N);
  }, [isFlying, N]);

  const next = useCallback(() => dismiss(1), [dismiss]);

  const onPointerDown = useCallback(
    (e: React.PointerEvent) => {
      if (isFlying) return;
      setDragging(true);
      startX.current = e.clientX;
      e.currentTarget.setPointerCapture(e.pointerId);
    },
    [isFlying],
  );

  const onPointerMove = useCallback(
    (e: React.PointerEvent) => {
      if (!dragging) return;
      setDragX(e.clientX - startX.current);
    },
    [dragging],
  );

  const onPointerUp = useCallback(() => {
    if (!dragging) return;
    setDragging(false);
    if (Math.abs(dragX) > 80) {
      dismiss(dragX < 0 ? 1 : -1);
    } else {
      setDragX(0);
    }
  }, [dragging, dragX, dismiss]);

  return (
    <section className="bg-secondary/60 py-16 md:py-24">
      {/* Heading */}
      <div className="mb-12 text-center">
        <h2 className="text-3xl font-normal sm:text-4xl">
          Reviews from <span className="font-semibold">real people</span>
        </h2>
        <div className="mt-4 flex items-center justify-center gap-2 text-sm text-muted-foreground">
          <Stars count={5} />
          <span className="font-semibold text-foreground">4.9/5</span>
          <span>·</span>
          <span>Based on early access feedback</span>
        </div>
      </div>

      <Container className="px-6">
        <div className="flex items-center gap-10">
          {/* Left panel */}
          <div className="hidden w-56 shrink-0 md:block">
            <svg
              viewBox="0 0 48 36"
              className="mb-5 size-10 fill-foreground/10"
              aria-hidden="true"
            >
              <path d="M0 36V22.5C0 10.5 7.5 3 22.5 0l3 4.5C17.5 6.5 13.5 11 13 18H22.5V36H0zm25.5 0V22.5C25.5 10.5 33 3 48 0l3 4.5C43 6.5 39 11 38.5 18H48V36H25.5z" />
            </svg>
            <h3 className="text-lg font-normal leading-snug text-foreground">
              What our customers are saying
            </h3>
            <div className="mt-8 flex items-center gap-3">
              <button
                onClick={prev}
                disabled={isFlying}
                className="flex size-8 items-center justify-center rounded-full border border-border text-muted-foreground transition-colors hover:border-foreground hover:text-foreground disabled:opacity-30"
              >
                <ChevronLeft className="size-4" />
              </button>
              <div className="h-px flex-1 bg-border" />
              <button
                onClick={next}
                disabled={isFlying}
                className="flex size-8 items-center justify-center rounded-full border border-border text-muted-foreground transition-colors hover:border-foreground hover:text-foreground disabled:opacity-30"
              >
                <ChevronRight className="size-4" />
              </button>
            </div>
            <div className="mt-5 flex gap-1.5">
              {REVIEWS.map((_, i) => (
                <button
                  key={i}
                  onClick={() => !isFlying && setCurrent(i)}
                  className={`h-1.5 rounded-full transition-all duration-300 ${
                    i === current
                      ? "w-6 bg-foreground"
                      : "w-1.5 bg-border hover:bg-muted-foreground"
                  }`}
                />
              ))}
            </div>
          </div>

          {/* Card stack */}
          <div className="relative flex-1" style={{ height: 300 }}>
            {REVIEWS.map((review, i) => {
              const depth = ((i - current) % N + N) % N;
              const isTop = depth === 0;
              const isHidden = depth > 2;

              let transform: string;
              if (isTop) {
                if (isFlying) {
                  transform = `translateX(${flyDir * 130}%) rotate(${flyDir * 18}deg)`;
                } else {
                  transform = `translateX(${dragX}px) rotate(${(dragX * 0.04).toFixed(2)}deg)`;
                }
              } else if (!isHidden) {
                transform = DEPTH_TRANSFORMS[depth];
              } else {
                transform = "translateY(30px) scale(0.88)";
              }

              return (
                <div
                  key={review.name}
                  className="absolute inset-x-0 top-0 touch-none select-none"
                  style={{
                    transform,
                    transition:
                      isTop && dragging
                        ? "none"
                        : "transform 0.32s ease, opacity 0.32s ease",
                    zIndex: isHidden ? 0 : N - depth,
                    opacity: isHidden ? 0 : 1,
                    cursor: isTop ? (dragging ? "grabbing" : "grab") : "default",
                    pointerEvents: isTop ? "auto" : "none",
                  }}
                  onPointerDown={isTop ? onPointerDown : undefined}
                  onPointerMove={isTop ? onPointerMove : undefined}
                  onPointerUp={isTop ? onPointerUp : undefined}
                  onPointerCancel={isTop ? onPointerUp : undefined}
                >
                  <ReviewCard review={review} />
                </div>
              );
            })}
          </div>
        </div>
      </Container>

      {/* Mobile nav */}
      <div className="mt-8 flex items-center justify-center gap-3 md:hidden">
        <button
          onClick={prev}
          className="flex size-8 items-center justify-center rounded-full border border-border text-muted-foreground"
        >
          <ChevronLeft className="size-4" />
        </button>
        <div className="flex gap-1.5">
          {REVIEWS.map((_, i) => (
            <button
              key={i}
              onClick={() => setCurrent(i)}
              className={`h-1.5 rounded-full transition-all duration-300 ${
                i === current ? "w-6 bg-foreground" : "w-1.5 bg-border"
              }`}
            />
          ))}
        </div>
        <button
          onClick={next}
          className="flex size-8 items-center justify-center rounded-full border border-border text-muted-foreground"
        >
          <ChevronRight className="size-4" />
        </button>
      </div>
    </section>
  );
}

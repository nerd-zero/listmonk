import { useState, useRef, useEffect, useCallback } from "react";
import { Container } from "@/components/landing/container";
import { MailingListsVisual } from "./mailing-lists-visual";
import { AnalyticsVisual } from "./analytics-visual";
import { TemplatingVisual } from "./templating-visual";
import { PerformanceVisual } from "./performance-visual";
import { TransactionalVisual } from "./transactional-visual";

const FEATURES = [
  {
    category: "Mailing lists",
    title: "Millions of subscribers, zero shared infrastructure.",
    description:
      "Single and double opt-in lists. Query and segment with raw SQL expressions.",
    Visual: MailingListsVisual,
  },
  {
    category: "Analytics",
    title: "See exactly how every campaign performs.",
    description:
      "Built-in open rates, click tracking, bounces, and top-link breakdowns across campaigns.",
    Visual: AnalyticsVisual,
  },
  {
    category: "Templating",
    title: "Go templates, drag-and-drop, Markdown, or plain HTML.",
    description:
      "Dynamic templates with 100+ built-in functions. Use logic in subject lines and body alike.",
    Visual: TemplatingVisual,
  },
  {
    category: "Performance",
    title: "7M+ e-mails on a fraction of a single core.",
    description:
      "Multi-threaded, multi-SMTP queues with sliding window rate limiting. 57 MB peak RAM.",
    Visual: PerformanceVisual,
  },
  {
    category: "Transactional",
    title: "One API call to reach any subscriber on any channel.",
    description:
      "Send e-mail, SMS, or WhatsApp messages via pre-defined templates and Messenger webhooks.",
    Visual: TransactionalVisual,
  },
];

export function Showcase() {
  const [current, setCurrent] = useState(0);
  const [isOverflowing, setIsOverflowing] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  const goTo = useCallback((idx: number) => {
    const container = scrollRef.current;
    if (!container) return;
    const card = container.querySelector<HTMLElement>(`[data-card="${idx}"]`);
    if (!card) return;
    const scrollLeft =
      container.scrollLeft +
      card.getBoundingClientRect().left -
      container.getBoundingClientRect().left;
    container.scrollTo({ left: scrollLeft, behavior: "smooth" });
    setCurrent(idx);
  }, []);

  useEffect(() => {
    const container = scrollRef.current;
    if (!container) return;
    const ro = new ResizeObserver(() => {
      setIsOverflowing(container.scrollWidth > container.clientWidth);
    });
    ro.observe(container);
    return () => ro.disconnect();
  }, []);

  useEffect(() => {
    const container = scrollRef.current;
    if (!container) return;
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.find(
          (e) => e.isIntersecting && e.intersectionRatio >= 0.5,
        );
        if (visible) {
          const idx = Number((visible.target as HTMLElement).dataset.card);
          setCurrent(idx);
        }
      },
      { root: container, threshold: 0.5 },
    );
    container
      .querySelectorAll("[data-card]")
      .forEach((el) => observer.observe(el));
    return () => observer.disconnect();
  }, []);

  return (
    <section id="showcase" className="pb-16 pt-6 md:pb-24 md:pt-8">
      <Container className="!max-w-screen-2xl">
        <div
          ref={scrollRef}
          className="flex gap-4 overflow-x-auto px-6 pb-2 [&::-webkit-scrollbar]:hidden"
          style={{ scrollbarWidth: "none" }}
        >
          {FEATURES.map((f, i) => (
            <div
              key={f.category}
              data-card={i}
              onClick={() => goTo(i)}
              className="flex min-w-72 flex-1 cursor-pointer flex-col gap-4 rounded-2xl bg-[#0d0d0d] p-5"
            >
              <div>
                <span className="text-[10px] font-semibold tracking-[0.14em] uppercase text-white/40">
                  {f.category}
                </span>
                <h3 className="mt-1.5 text-sm font-normal leading-snug text-white">
                  {f.title}
                </h3>
                <p className="mt-1.5 text-[11px] text-white/50 line-clamp-2">
                  {f.description}
                </p>
              </div>
              <div className="rounded-lg border border-white/10 bg-black/40 p-3">
                <f.Visual />
              </div>
            </div>
          ))}
        </div>

        {isOverflowing && (
          <div className="mt-6 flex justify-center">
            <div className="flex items-center gap-1.5">
              {FEATURES.map((_, i) => (
                <button
                  key={i}
                  onClick={() => goTo(i)}
                  aria-label={`Go to slide ${i + 1}`}
                  className={`h-1.5 rounded-full transition-all duration-300 ${
                    i === current
                      ? "w-6 bg-foreground"
                      : "w-1.5 bg-border hover:bg-muted-foreground"
                  }`}
                />
              ))}
            </div>
          </div>
        )}
      </Container>
    </section>
  );
}

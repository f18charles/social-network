import { Link } from "react-router";
import { useAuth } from "../context/auth/useAuth.js";
import { Navigate } from "react-router";
import "../styles/welcome.css";

const FEATURES = [
  {
    icon: "✦",
    title: "Share Your World",
    desc: "Post thoughts, photos, and moments with the people who matter. Control exactly who sees what.",
    color: "cyan",
  },
  {
    icon: "⬡",
    title: "Build Your Circle",
    desc: "Follow friends, accept followers, and nurture your network — all with built-in privacy controls.",
    color: "violet",
  },
  {
    icon: "◈",
    title: "Groups & Events",
    desc: "Create communities around shared interests and organise events your crew actually shows up to.",
    color: "green",
  },
  {
    icon: "◎",
    title: "Real-Time Chat",
    desc: "DM anyone or jump into group threads. Messages land instantly — no refresh required.",
    color: "pink",
  },
];

function Welcome() {
  const { isAuthenticated, isLoading } = useAuth();

  // Redirect authenticated users straight to the feed
  if (!isLoading && isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return (
    <div className="welcome-page">
      {/* ── Animated grid background ── */}
      <div className="welcome-grid" aria-hidden="true" />

      {/* ── Floating orbs ── */}
      <div className="welcome-orb welcome-orb--violet" aria-hidden="true" />
      <div className="welcome-orb welcome-orb--cyan" aria-hidden="true" />
      <div className="welcome-orb welcome-orb--green" aria-hidden="true" />

      {/* ── Nav bar ── */}
      <header className="welcome-nav">
        <span className="welcome-nav__logo">
          <span className="welcome-nav__logo-icon">◈</span>
          <span className="welcome-nav__logo-text">SocialNet</span>
        </span>
        <nav className="welcome-nav__links">
          <Link to="/login" className="welcome-nav__link">
            Log in
          </Link>
          <Link to="/register" className="welcome-nav__cta">
            Get started
          </Link>
        </nav>
      </header>

      {/* ── Hero ── */}
      <section className="welcome-hero">
        <div className="welcome-hero__badge">
          <span className="welcome-hero__badge-dot" />
          Now live — join thousands already connected
        </div>

        <h1 className="welcome-hero__headline">
          Your network,
          <br />
          <span className="welcome-hero__headline--gradient">your rules.</span>
        </h1>

        <p className="welcome-hero__sub">
          A social platform built around authentic connection — not algorithms.
          Share posts, join groups, chat in real time, and control exactly who
          sees your life.
        </p>

        <div className="welcome-hero__actions">
          <Link to="/register" className="welcome-btn welcome-btn--primary">
            Create free account
          </Link>
          <Link to="/login" className="welcome-btn welcome-btn--ghost">
            Sign in
          </Link>
        </div>

        {/* stat strip */}
        <div className="welcome-stats">
          {[
            { value: "10K+", label: "Members" },
            { value: "50K+", label: "Posts shared" },
            { value: "99.9%", label: "Uptime" },
          ].map(({ value, label }) => (
            <div key={label} className="welcome-stat">
              <span className="welcome-stat__value">{value}</span>
              <span className="welcome-stat__label">{label}</span>
            </div>
          ))}
        </div>
      </section>

      {/* ── Features ── */}
      <section className="welcome-features">
        <h2 className="welcome-section-title">
          Everything you need,{" "}
          <span className="welcome-section-title--accent">nothing you don't.</span>
        </h2>

        <div className="welcome-feature-grid">
          {FEATURES.map(({ icon, title, desc, color }) => (
            <article
              key={title}
              className={`welcome-feature-card welcome-feature-card--${color}`}
            >
              <span className="welcome-feature-card__icon">{icon}</span>
              <h3 className="welcome-feature-card__title">{title}</h3>
              <p className="welcome-feature-card__desc">{desc}</p>
            </article>
          ))}
        </div>
      </section>

      {/* ── CTA banner ── */}
      <section className="welcome-cta-banner">
        <div className="welcome-cta-banner__inner">
          <h2 className="welcome-cta-banner__title">Ready to connect?</h2>
          <p className="welcome-cta-banner__sub">
            Free forever. No ads. No data selling. Just people.
          </p>
          <Link to="/register" className="welcome-btn welcome-btn--primary welcome-btn--lg">
            Join SocialNet →
          </Link>
        </div>
      </section>

      {/* ── Footer ── */}
      <footer className="welcome-footer">
        <span className="welcome-footer__brand">
          <span className="welcome-footer__brand-icon">◈</span> SocialNet
        </span>
        <span className="welcome-footer__copy">
          Already a member?{" "}
          <Link to="/login" className="welcome-footer__link">
            Log in here
          </Link>
        </span>
      </footer>
    </div>
  );
}

export default Welcome;

<script setup lang="ts">
// DataBuddy 官方风格的 3D 磨砂质感专家团插画 + 全息数据卡片 + 光效背景
// 纯 SVG 绘制，零外部资源。适合登录页与 Buddy 首页复用。
defineProps<{ compact?: boolean }>()
</script>

<template>
  <svg class="team-hero" :class="{ compact }" viewBox="0 0 1000 340" preserveAspectRatio="xMidYMax meet" role="img" aria-label="数据工程与治理专家团">
    <defs>
      <!-- 背景光效 -->
      <radialGradient id="th-glow-a" cx="50%" cy="50%" r="50%">
        <stop offset="0%" stop-color="#b0d4f6" stop-opacity="0.55" />
        <stop offset="100%" stop-color="#b0d4f6" stop-opacity="0" />
      </radialGradient>
      <radialGradient id="th-glow-b" cx="50%" cy="50%" r="50%">
        <stop offset="0%" stop-color="#c8b4f0" stop-opacity="0.32" />
        <stop offset="100%" stop-color="#c8b4f0" stop-opacity="0" />
      </radialGradient>
      <radialGradient id="th-glow-c" cx="50%" cy="50%" r="50%">
        <stop offset="0%" stop-color="#8fe3d8" stop-opacity="0.3" />
        <stop offset="100%" stop-color="#8fe3d8" stop-opacity="0" />
      </radialGradient>
      <pattern id="th-dots" width="22" height="22" patternUnits="userSpaceOnUse">
        <circle cx="2" cy="2" r="1.1" fill="#5a8fc4" opacity="0.11" />
      </pattern>
      <!-- 皮肤 / 衣服渐变 -->
      <radialGradient id="th-skin" cx="38%" cy="30%" r="75%">
        <stop offset="0%" stop-color="#ffe0c8" /><stop offset="65%" stop-color="#f3c4a0" /><stop offset="100%" stop-color="#d9a17e" />
      </radialGradient>
      <radialGradient id="th-skin2" cx="38%" cy="30%" r="75%">
        <stop offset="0%" stop-color="#ffdfc6" /><stop offset="65%" stop-color="#eeb994" /><stop offset="100%" stop-color="#cf9570" />
      </radialGradient>
      <linearGradient id="th-vest" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#5a6779" /><stop offset="100%" stop-color="#3a4557" />
      </linearGradient>
      <linearGradient id="th-shirt" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#f2f4f7" /><stop offset="100%" stop-color="#d9dee5" />
      </linearGradient>
      <linearGradient id="th-sweater" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#e2d3bd" /><stop offset="100%" stop-color="#c9b79b" />
      </linearGradient>
      <linearGradient id="th-polo" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#adb6c2" /><stop offset="100%" stop-color="#8b95a3" />
      </linearGradient>
      <linearGradient id="th-whitet" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#ffffff" /><stop offset="100%" stop-color="#e6e9ee" />
      </linearGradient>
      <linearGradient id="th-denim" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#2e3b50" /><stop offset="100%" stop-color="#1a2433" />
      </linearGradient>
      <linearGradient id="th-hair-red" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#c2682f" /><stop offset="100%" stop-color="#9c4a1e" />
      </linearGradient>
      <linearGradient id="th-hair-dark" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#3a3f48" /><stop offset="100%" stop-color="#21252c" />
      </linearGradient>
      <linearGradient id="th-hair-blond" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#e8c98f" /><stop offset="100%" stop-color="#c9a461" />
      </linearGradient>
      <linearGradient id="th-hair-gray" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#d6dae0" /><stop offset="100%" stop-color="#aeb5bf" />
      </linearGradient>
      <linearGradient id="th-card" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#ffffff" stop-opacity="0.95" />
        <stop offset="100%" stop-color="#eaf2ff" stop-opacity="0.72" />
      </linearGradient>
      <linearGradient id="th-star" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#60a5fa" /><stop offset="100%" stop-color="#34d399" />
      </linearGradient>
      <filter id="th-shadow" x="-40%" y="-40%" width="180%" height="180%">
        <feGaussianBlur stdDeviation="9" />
      </filter>
      <filter id="th-ground" x="-60%" y="-200%" width="220%" height="500%">
        <feGaussianBlur stdDeviation="11" />
      </filter>
      <!-- 统一左上打光：柔和高光 -->
      <radialGradient id="th-light" cx="28%" cy="18%" r="70%">
        <stop offset="0%" stop-color="#ffffff" stop-opacity="0.34" />
        <stop offset="100%" stop-color="#ffffff" stop-opacity="0" />
      </radialGradient>
      <filter id="th-soft" x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow dx="0" dy="5" stdDeviation="6" flood-color="#24405f" flood-opacity="0.16" />
      </filter>
    </defs>

    <!-- 光斑 + 点阵 -->
    <ellipse cx="500" cy="205" rx="430" ry="150" fill="url(#th-glow-a)" />
    <ellipse cx="250" cy="150" rx="230" ry="140" fill="url(#th-glow-b)" />
    <ellipse cx="790" cy="240" rx="250" ry="120" fill="url(#th-glow-c)" />
    <rect x="0" y="0" width="1000" height="340" fill="url(#th-dots)" />
    <ellipse cx="500" cy="322" rx="360" ry="14" fill="#6f9fd0" opacity="0.18" filter="url(#th-shadow)" />

    <!-- 人物组（统一柔和投影） -->
    <g filter="url(#th-soft)">
    <!-- 人物 1：白发男，深蓝灰背心 -->
    <g transform="translate(178,0)">
      <ellipse cx="0" cy="316" rx="54" ry="13" fill="#2c3f57" opacity="0.22" filter="url(#th-ground)" />
      <path d="M-46 316 C-46 250 -30 222 0 222 C30 222 46 250 46 316 Z" fill="url(#th-vest)" />
      <path d="M-22 226 C-22 210 22 210 22 226 L18 248 C8 256 -8 256 -18 248 Z" fill="url(#th-shirt)" />
      <rect x="-9" y="206" width="18" height="22" rx="8" fill="url(#th-skin)" />
      <ellipse cx="0" cy="182" rx="30" ry="34" fill="url(#th-skin)" />
      <path d="M-30 178 C-32 156 -18 146 0 146 C20 146 32 158 29 178 C24 166 14 162 0 162 C-14 162 -24 166 -30 178 Z" fill="url(#th-hair-gray)" />
      <ellipse cx="-11" cy="182" rx="2.6" ry="3" fill="#3b4653" /><ellipse cx="11" cy="182" rx="2.6" ry="3" fill="#3b4653" />
      <path d="M-9 198 Q0 204 9 198" stroke="#a86d54" stroke-width="2" fill="none" stroke-linecap="round" />
    </g>

    <!-- 人物 2：红发女，米色毛衣 -->
    <g transform="translate(348,0)">
      <ellipse cx="0" cy="316" rx="56" ry="13" fill="#2c3f57" opacity="0.22" filter="url(#th-ground)" />
      <path d="M-48 316 C-50 246 -30 218 0 218 C30 218 50 246 48 316 Z" fill="url(#th-sweater)" />
      <rect x="-9" y="202" width="18" height="22" rx="8" fill="url(#th-skin2)" />
      <ellipse cx="0" cy="176" rx="31" ry="35" fill="url(#th-skin2)" />
      <path d="M-33 176 C-38 146 -20 136 2 136 C26 136 36 152 32 182 C28 164 18 158 2 158 C-14 158 -26 162 -33 176 Z" fill="url(#th-hair-red)" />
      <path d="M-33 176 C-38 196 -32 214 -26 220 C-30 200 -28 184 -26 174 Z" fill="url(#th-hair-red)" />
      <ellipse cx="-11" cy="178" rx="2.6" ry="3" fill="#4a3730" /><ellipse cx="11" cy="178" rx="2.6" ry="3" fill="#4a3730" />
      <path d="M-8 193 Q0 199 8 193" stroke="#b26a55" stroke-width="2" fill="none" stroke-linecap="round" />
    </g>

    <!-- 人物 3（C位）：黑发眼镜男，灰色 Polo -->
    <g transform="translate(500,0)">
      <ellipse cx="0" cy="316" rx="60" ry="14" fill="#2c3f57" opacity="0.24" filter="url(#th-ground)" />
      <path d="M-52 316 C-54 240 -32 210 0 210 C32 210 54 240 52 316 Z" fill="url(#th-polo)" />
      <path d="M-14 211 L0 230 L14 211 L8 208 L0 220 L-8 208 Z" fill="#6f7a89" />
      <rect x="-10" y="194" width="20" height="24" rx="8" fill="url(#th-skin)" />
      <ellipse cx="0" cy="166" rx="33" ry="37" fill="url(#th-skin)" />
      <path d="M-34 162 C-38 136 -20 126 2 126 C26 126 36 142 33 166 C28 150 16 146 0 146 C-16 146 -28 150 -34 162 Z" fill="url(#th-hair-dark)" />
      <!-- 眼镜 -->
      <g stroke="#2b2f36" stroke-width="2.4" fill="none">
        <rect x="-23" y="162" width="19" height="14" rx="5" /><rect x="4" y="162" width="19" height="14" rx="5" />
        <path d="M-4 168 H4" /><path d="M-23 167 L-32 164" /><path d="M23 167 L32 164" />
      </g>
      <ellipse cx="-13.5" cy="169" rx="2.4" ry="2.6" fill="#3b4653" /><ellipse cx="13.5" cy="169" rx="2.4" ry="2.6" fill="#3b4653" />
      <path d="M-9 186 Q0 192 9 186" stroke="#a86d54" stroke-width="2" fill="none" stroke-linecap="round" />
    </g>

    <!-- 人物 4：金发女，白T -->
    <g transform="translate(652,0)">
      <ellipse cx="0" cy="316" rx="56" ry="13" fill="#2c3f57" opacity="0.22" filter="url(#th-ground)" />
      <path d="M-46 316 C-48 246 -28 218 0 218 C28 218 48 246 46 316 Z" fill="url(#th-whitet)" />
      <rect x="-9" y="202" width="18" height="22" rx="8" fill="url(#th-skin2)" />
      <ellipse cx="0" cy="176" rx="31" ry="35" fill="url(#th-skin2)" />
      <path d="M-32 172 C-36 146 -18 134 2 134 C25 134 35 150 33 174 C29 154 16 148 0 148 C-15 148 -27 154 -32 172 Z" fill="url(#th-hair-blond)" />
      <path d="M-32 172 C-38 192 -34 210 -28 216 C-32 196 -30 180 -28 170 Z" fill="url(#th-hair-blond)" />
      <path d="M33 174 C38 194 34 212 28 218 C32 198 30 182 28 172 Z" fill="url(#th-hair-blond)" />
      <ellipse cx="-11" cy="178" rx="2.6" ry="3" fill="#4a5a63" /><ellipse cx="11" cy="178" rx="2.6" ry="3" fill="#4a5a63" />
      <path d="M-8 193 Q0 199 8 193" stroke="#b26a55" stroke-width="2" fill="none" stroke-linecap="round" />
    </g>

    <!-- 人物 5：黑发男，深藏青夹克 -->
    <g transform="translate(822,0)">
      <ellipse cx="0" cy="316" rx="54" ry="13" fill="#2c3f57" opacity="0.22" filter="url(#th-ground)" />
      <path d="M-46 316 C-48 250 -30 222 0 222 C30 222 48 250 46 316 Z" fill="url(#th-denim)" />
      <path d="M-14 224 L0 244 L14 224" stroke="#455368" stroke-width="3" fill="none" />
      <rect x="-9" y="206" width="18" height="22" rx="8" fill="url(#th-skin)" />
      <ellipse cx="0" cy="182" rx="30" ry="34" fill="url(#th-skin)" />
      <path d="M-31 178 C-34 154 -18 144 2 144 C24 144 34 158 30 180 C26 164 14 160 0 160 C-14 160 -26 164 -31 178 Z" fill="url(#th-hair-dark)" />
      <ellipse cx="-11" cy="182" rx="2.6" ry="3" fill="#3b4653" /><ellipse cx="11" cy="182" rx="2.6" ry="3" fill="#3b4653" />
      <path d="M-9 197 Q0 203 9 197" stroke="#a86d54" stroke-width="2" fill="none" stroke-linecap="round" />
    </g>
    </g>
    <!-- 统一高光罩 -->
    <ellipse cx="500" cy="180" rx="420" ry="150" fill="url(#th-light)" pointer-events="none" />

    <!-- 全息数据卡片 -->
    <g>
      <rect x="18" y="92" width="150" height="86" rx="12" fill="url(#th-card)" stroke="#cfe0f5" />
      <g stroke="#8fb6e2" stroke-width="3" stroke-linecap="round">
        <path d="M38 118h110" /><path d="M38 134h78" /><path d="M38 150h96" />
      </g>
      <circle cx="148" cy="104" r="4" fill="#34d399" />
    </g>
    <g>
      <rect x="832" y="86" width="150" height="96" rx="12" fill="url(#th-card)" stroke="#cfe0f5" />
      <circle cx="907" cy="134" r="28" fill="none" stroke="#9dbdea" stroke-width="10" stroke-dasharray="112 64" stroke-linecap="round" transform="rotate(-90 907 134)" />
      <circle cx="907" cy="134" r="10" fill="#00a2b3" opacity="0.85" />
    </g>
    <g opacity="0.72">
      <rect x="788" y="214" width="120" height="76" rx="14" fill="url(#th-card)" stroke="#dbe9fb" stroke-opacity="0.7" />
      <g fill="#a9cef2" opacity="0.85">
        <rect x="806" y="258" width="11" height="20" rx="5" /><rect x="825" y="244" width="11" height="34" rx="5" />
        <rect x="844" y="250" width="11" height="28" rx="5" /><rect x="863" y="236" width="11" height="42" rx="5" />
      </g>
    </g>
    <g opacity="0.72">
      <rect x="62" y="210" width="120" height="76" rx="14" fill="url(#th-card)" stroke="#dbe9fb" stroke-opacity="0.7" />
      <path d="M78 262 q8 -9 16 -9 t16 -9 t20 -9" stroke="#9cc5ee" stroke-width="3.5" fill="none" stroke-linecap="round" />
      <circle cx="130" cy="235" r="4" fill="#00a2b3" opacity="0.8" />
      <circle cx="130" cy="235" r="8" fill="none" stroke="#00a2b3" stroke-opacity="0.28" stroke-width="2" />
    </g>

    <!-- 星形装饰 -->
    <path d="M500 44 l4 10 10 4 -10 4 -4 10 -4 -10 -10 -4 10 -4 z" fill="url(#th-star)" opacity="0.9" />
    <path d="M120 40 l3 7 7 3 -7 3 -3 7 -3 -7 -7 -3 7 -3 z" fill="url(#th-star)" opacity="0.7" />
    <path d="M900 46 l3 7 7 3 -7 3 -3 7 -3 -7 -7 -3 7 -3 z" fill="url(#th-star)" opacity="0.7" />
  </svg>
</template>

<style scoped>
.team-hero { width: 100%; height: auto; display: block; }
.team-hero.compact { max-height: 260px; }
</style>

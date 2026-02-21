// Willow Documentation — Navigation Data & Sidebar Logic

const REFERENCE_ITEMS = [
    { label: "API Reference", href: "https://pkg.go.dev/github.com/phanxgames/willow" },
    { label: "GitHub", href: "https://github.com/phanxgames/willow" },
];

const NAV_TABS = [
    { id: "docs", label: "Docs", sections: [
        {
            title: "Getting Started",
            items: [
                { label: "Getting Started", page: "getting-started" },
                { label: "Architecture", page: "architecture" },
            ]
        },
        {
            title: "Core",
            items: [
                { label: "Scene", page: "scene" },
                { label: "Nodes", page: "nodes" },
                { label: "Transforms", page: "transforms" },
            ]
        },
        {
            title: "Rendering",
            items: [
                { label: "Sprites & Atlas", page: "sprites-and-atlas" },
                { label: "Camera & Viewport", page: "camera-and-viewport" },
                { label: "Text & Fonts", page: "text-and-fonts" },
                { label: "Tilemap Viewport", page: "tilemap-viewport" },
            ]
        },
        {
            title: "Input",
            items: [
                { label: "Input, Hit Testing & Gestures", page: "input-hit-testing-and-gestures" },
            ]
        },
        {
            title: "Effects",
            items: [
                { label: "Tweens & Animation", page: "tweens-and-animation" },
                { label: "Particles", page: "particles" },
                { label: "Lighting", page: "lighting" },
                { label: "Post-Processing Filters", page: "post-processing-filters" },
            ]
        },
        {
            title: "Advanced",
            items: [
                { label: "Mesh, Ropes & Polygons", page: "mesh-ropes-and-polygons" },
                { label: "Offscreen Rendering", page: "offscreen-rendering" },
                { label: "Clipping & Masks", page: "clipping-and-masks" },
                { label: "Performance Caching", page: "performance-caching" },
            ]
        },
        {
            title: "Tools",
            items: [
                { label: "Debug & Testing", page: "debug-and-testing" },
                { label: "ECS Integration", page: "ecs-integration" },
            ]
        },
    ]},
    { id: "examples", label: "Examples", sections: [
        { title: "Examples", items: [{ label: "Examples", page: "examples" }] }
    ]},
    { id: "demos", label: "Demos", sections: [
        { title: "Demos", items: [{ label: "Demos (Coming Soon)", page: "demos" }] }
    ]},
    { id: "tutorials", label: "Tutorials", sections: [
        { title: "Tutorials", items: [{ label: "Tutorials (Coming Soon)", page: "tutorials" }] }
    ]},
    { id: "reference", label: "Reference", sections: [
        { title: "Reference", items: REFERENCE_ITEMS }
    ]}
];

let activeTabId = "docs";

function getActivePage() {
    const params = new URLSearchParams(window.location.search);
    return params.get("page") || "getting-started";
}

function findTabForPage(page) {
    for (const tab of NAV_TABS) {
        for (const section of tab.sections) {
            for (const item of section.items) {
                if (item.page === page) return tab.id;
            }
        }
    }
    return "docs";
}

function buildTabs() {
    const container = document.getElementById("sidebar-tabs");
    container.innerHTML = "";

    NAV_TABS.forEach(tab => {
        const btn = document.createElement("button");
        btn.className = "sidebar-tab" + (tab.id === activeTabId ? " active" : "");
        btn.textContent = tab.label;
        btn.addEventListener("click", () => switchTab(tab.id));
        container.appendChild(btn);
    });
}

function switchTab(tabId) {
    activeTabId = tabId;

    document.querySelectorAll(".sidebar-tab").forEach(btn => {
        btn.classList.toggle("active", btn.textContent === NAV_TABS.find(t => t.id === tabId).label);
    });

    const tab = NAV_TABS.find(t => t.id === tabId);
    if (tab) buildSidebar(tab.sections);
}

function buildFooter() {
    const footer = document.getElementById("sidebar-footer");
    footer.innerHTML = "";

    REFERENCE_ITEMS.forEach(item => {
        const a = document.createElement("a");
        a.className = "sidebar-footer-link";
        a.textContent = item.label;
        a.href = item.href;
        a.target = "_blank";
        a.rel = "noopener noreferrer";
        footer.appendChild(a);
    });
}

function buildSidebar(sections) {
    const nav = document.getElementById("sidebar-nav");
    const activePage = getActivePage();
    nav.innerHTML = "";

    sections.forEach(section => {
        const sectionEl = document.createElement("div");
        sectionEl.className = "nav-section";

        const titleEl = document.createElement("div");
        titleEl.className = "nav-section-title";
        titleEl.innerHTML = `${section.title} <span class="chevron">&#9662;</span>`;
        titleEl.addEventListener("click", () => {
            sectionEl.classList.toggle("collapsed");
        });
        sectionEl.appendChild(titleEl);

        const listEl = document.createElement("ul");
        listEl.className = "nav-items";

        section.items.forEach(item => {
            const li = document.createElement("li");
            li.className = "nav-item";
            const a = document.createElement("a");
            a.textContent = item.label;
            if (item.href) {
                a.href = item.href;
                a.target = "_blank";
                a.rel = "noopener noreferrer";
            } else {
                a.href = `?page=${item.page}`;
                if (item.page === activePage) {
                    a.className = "active";
                }
                a.addEventListener("click", (e) => {
                    e.preventDefault();
                    navigateTo(item.page);
                });
            }
            li.appendChild(a);
            listEl.appendChild(li);
        });

        sectionEl.appendChild(listEl);
        nav.appendChild(sectionEl);
    });
}

function navigateTo(page) {
    if (!page) return;
    history.pushState({page}, "", `?page=${page}`);

    // Switch to correct tab if needed
    const tabId = findTabForPage(page);
    if (tabId !== activeTabId) {
        switchTab(tabId);
    }

    // Update active state in sidebar
    document.querySelectorAll(".nav-item a").forEach(a => {
        const href = a.getAttribute("href");
        a.classList.toggle("active", href === `?page=${page}`);
    });

    // Load page in iframe
    const iframe = document.getElementById("content-frame");
    iframe.src = `viewer.html?page=${page}`;

    // Close mobile sidebar
    document.querySelector(".sidebar").classList.remove("open");
    document.querySelector(".overlay").classList.remove("visible");
}

window.addEventListener("popstate", (e) => {
    const page = (e.state && e.state.page) || getActivePage();

    const tabId = findTabForPage(page);
    if (tabId !== activeTabId) {
        switchTab(tabId);
    }

    const iframe = document.getElementById("content-frame");
    iframe.src = `viewer.html?page=${page}`;

    document.querySelectorAll(".nav-item a").forEach(a => {
        const href = a.getAttribute("href");
        a.classList.toggle("active", href === `?page=${page}`);
    });
});

document.addEventListener("DOMContentLoaded", () => {
    const activePage = getActivePage();
    activeTabId = findTabForPage(activePage);

    buildTabs();
    buildFooter();
    switchTab(activeTabId);

    // Hamburger toggle
    const hamburger = document.querySelector(".hamburger");
    const sidebar = document.querySelector(".sidebar");
    const overlay = document.querySelector(".overlay");

    if (hamburger) {
        hamburger.addEventListener("click", () => {
            sidebar.classList.toggle("open");
            overlay.classList.toggle("visible");
        });
    }

    if (overlay) {
        overlay.addEventListener("click", () => {
            sidebar.classList.remove("open");
            overlay.classList.remove("visible");
        });
    }

    // Load initial page in iframe
    const iframe = document.getElementById("content-frame");
    iframe.src = `viewer.html?page=${activePage}`;
});

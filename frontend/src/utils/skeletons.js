// RIFT Premium Dark Skeleton Loaders for all sections

export function getProfileSkeletonHTML() {
    return `
        <div class="view-skeleton fade-in" style="max-width: 1140px; margin: 0 auto; padding: 36px 44px; padding-bottom: 80px;">
            <!-- Top Identity Card Skeleton -->
            <div style="background: #0d0f17; border: 1px solid rgba(255,255,255,0.06); border-radius: 6px; padding: 32px 36px; margin-bottom: 32px; display: flex; justify-content: space-between; align-items: center; gap: 28px; flex-wrap: wrap;">
                <div style="display: flex; align-items: center; gap: 26px; flex-wrap: wrap;">
                    <div class="skeleton-shimmer" style="width: 84px; height: 84px; border-radius: 50%; flex-shrink: 0;"></div>
                    <div>
                        <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">
                            <div class="skeleton-shimmer" style="width: 180px; height: 28px; border-radius: 4px;"></div>
                            <div class="skeleton-shimmer" style="width: 110px; height: 22px; border-radius: 4px;"></div>
                        </div>
                        <div class="skeleton-shimmer" style="width: 140px; height: 14px; border-radius: 4px; margin-bottom: 14px;"></div>
                        <div style="display: flex; gap: 8px;">
                            <div class="skeleton-shimmer" style="width: 90px; height: 24px; border-radius: 4px;"></div>
                            <div class="skeleton-shimmer" style="width: 110px; height: 24px; border-radius: 4px;"></div>
                            <div class="skeleton-shimmer" style="width: 120px; height: 24px; border-radius: 4px;"></div>
                        </div>
                    </div>
                </div>
                <div style="border-left: 1px solid rgba(255,255,255,0.06); padding-left: 24px;">
                    <div class="skeleton-shimmer" style="width: 140px; height: 22px; border-radius: 4px; margin-bottom: 8px;"></div>
                    <div class="skeleton-shimmer" style="width: 90px; height: 14px; border-radius: 4px;"></div>
                </div>
            </div>

            <!-- Retrospective 4 Stat Cards Skeleton -->
            <div style="display: flex; justify-content: space-between; margin: 36px 0 16px 0;">
                <div class="skeleton-shimmer" style="width: 220px; height: 24px; border-radius: 4px;"></div>
                <div class="skeleton-shimmer" style="width: 100px; height: 16px; border-radius: 4px;"></div>
            </div>
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; margin-bottom: 36px;">
                ${[1,2,3,4].map(() => `
                    <div style="background: #0d0f17; border: 1px solid rgba(255,255,255,0.06); border-radius: 6px; padding: 22px 24px; min-height: 128px; display: flex; flex-direction: column; justify-content: space-between;">
                        <div style="display: flex; justify-content: space-between; align-items: center;">
                            <div class="skeleton-shimmer" style="width: 80px; height: 14px;"></div>
                            <div class="skeleton-shimmer" style="width: 50px; height: 16px;"></div>
                        </div>
                        <div>
                            <div class="skeleton-shimmer" style="width: 100px; height: 38px; margin-bottom: 6px;"></div>
                            <div class="skeleton-shimmer" style="width: 120px; height: 14px;"></div>
                        </div>
                    </div>
                `).join('')}
            </div>

            <!-- Top 4 Games Section Skeleton -->
            <div style="display: flex; justify-content: space-between; margin: 36px 0 16px 0;">
                <div class="skeleton-shimmer" style="width: 200px; height: 24px; border-radius: 4px;"></div>
                <div class="skeleton-shimmer" style="width: 90px; height: 16px; border-radius: 4px;"></div>
            </div>
            <div style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 22px; margin-bottom: 44px;">
                ${[1,2,3,4].map(() => `
                    <div>
                        <div class="skeleton-shimmer" style="width: 100%; aspect-ratio: 10 / 14; border-radius: 8px; margin-bottom: 12px;"></div>
                        <div class="skeleton-shimmer" style="width: 85%; height: 18px; margin-bottom: 6px;"></div>
                        <div class="skeleton-shimmer" style="width: 45%; height: 14px;"></div>
                    </div>
                `).join('')}
            </div>

            <!-- Bulletins Box Skeleton -->
            <div style="display: flex; justify-content: space-between; margin: 36px 0 16px 0;">
                <div class="skeleton-shimmer" style="width: 230px; height: 24px; border-radius: 4px;"></div>
                <div class="skeleton-shimmer" style="width: 120px; height: 20px; border-radius: 4px;"></div>
            </div>
            <div class="skeleton-shimmer" style="width: 100%; height: 100px; border-radius: 6px; margin-bottom: 44px;"></div>

            <!-- 2 Arcade Cards Skeleton -->
            <div style="display: flex; justify-content: space-between; margin: 36px 0 16px 0;">
                <div class="skeleton-shimmer" style="width: 210px; height: 24px; border-radius: 4px;"></div>
                <div class="skeleton-shimmer" style="width: 130px; height: 16px; border-radius: 4px;"></div>
            </div>
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(310px, 1fr)); gap: 20px;">
                <div class="skeleton-shimmer" style="height: 180px; border-radius: 8px;"></div>
                <div class="skeleton-shimmer" style="height: 180px; border-radius: 8px;"></div>
            </div>
        </div>
    `;
}

export function getLibrarySkeletonHTML() {
    return `
        <div class="view-skeleton fade-in" style="padding: 28px 40px; height: 100%; overflow-y: auto;">
            <!-- Hero Banner Skeleton -->
            <div class="skeleton-shimmer" style="width: 100%; height: 260px; border-radius: 12px; margin-bottom: 32px;"></div>

            <!-- Controls / Filter Tabs Skeleton -->
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; gap: 16px;">
                <div style="display: flex; gap: 10px;">
                    <div class="skeleton-shimmer" style="width: 70px; height: 32px; border-radius: 6px;"></div>
                    <div class="skeleton-shimmer" style="width: 80px; height: 32px; border-radius: 6px;"></div>
                    <div class="skeleton-shimmer" style="width: 80px; height: 32px; border-radius: 6px;"></div>
                </div>
                <div class="skeleton-shimmer" style="width: 240px; height: 34px; border-radius: 6px;"></div>
            </div>

            <!-- Grid of 8 Posters Skeleton -->
            <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 24px;">
                ${[1,2,3,4,5,6,7,8].map(() => `
                    <div>
                        <div class="skeleton-shimmer" style="width: 100%; aspect-ratio: 10 / 14; border-radius: 8px; margin-bottom: 10px;"></div>
                        <div class="skeleton-shimmer" style="width: 80%; height: 16px; margin-bottom: 6px;"></div>
                        <div class="skeleton-shimmer" style="width: 50%; height: 12px;"></div>
                    </div>
                `).join('')}
            </div>
        </div>
    `;
}

export function getStoreSkeletonHTML() {
    return `
        <div class="view-skeleton fade-in" style="padding: 28px 40px; height: 100%; overflow-y: auto;">
            <!-- Storefront Banners Skeleton -->
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 18px; margin-bottom: 36px;">
                <div class="skeleton-shimmer" style="height: 100px; border-radius: 10px;"></div>
                <div class="skeleton-shimmer" style="height: 100px; border-radius: 10px;"></div>
            </div>

            <!-- Section Title -->
            <div style="display: flex; justify-content: space-between; margin-bottom: 20px;">
                <div class="skeleton-shimmer" style="width: 220px; height: 26px; border-radius: 4px;"></div>
                <div class="skeleton-shimmer" style="width: 100px; height: 20px; border-radius: 4px;"></div>
            </div>

            <!-- Store Cards Grid -->
            <div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 20px;">
                ${[1,2,3,4,5,6].map(() => `
                    <div>
                        <div class="skeleton-shimmer" style="width: 100%; height: 140px; border-radius: 8px; margin-bottom: 10px;"></div>
                        <div class="skeleton-shimmer" style="width: 75%; height: 16px; margin-bottom: 6px;"></div>
                        <div class="skeleton-shimmer" style="width: 40%; height: 14px;"></div>
                    </div>
                `).join('')}
            </div>
        </div>
    `;
}

export function getGameDetailSkeletonHTML() {
    return `
        <div class="view-skeleton fade-in" style="padding: 28px 40px; height: 100%; overflow-y: auto;">
            <div class="skeleton-shimmer" style="width: 100%; height: 320px; border-radius: 12px; margin-bottom: 30px;"></div>
            <div style="display: grid; grid-template-columns: 2fr 1fr; gap: 32px;">
                <div>
                    <div class="skeleton-shimmer" style="width: 60%; height: 36px; border-radius: 4px; margin-bottom: 14px;"></div>
                    <div style="display: flex; gap: 10px; margin-bottom: 24px;">
                        <div class="skeleton-shimmer" style="width: 90px; height: 24px; border-radius: 4px;"></div>
                        <div class="skeleton-shimmer" style="width: 120px; height: 24px; border-radius: 4px;"></div>
                    </div>
                    <div class="skeleton-shimmer" style="width: 100%; height: 120px; border-radius: 8px;"></div>
                </div>
                <div>
                    <div class="skeleton-shimmer" style="width: 100%; height: 240px; border-radius: 8px;"></div>
                </div>
            </div>
        </div>
    `;
}

export function getGenericSkeletonHTML() {
    // Elegant dark card grid matching the user's uploaded mockup
    return `
        <div class="view-skeleton fade-in" style="max-width: 960px; margin: 0 auto; padding: 32px 36px; height: 100%; overflow-y: auto;">
            <!-- Top circle avatar + pill -->
            <div style="display: flex; align-items: center; gap: 16px; margin-bottom: 30px;">
                <div class="skeleton-shimmer" style="width: 52px; height: 52px; border-radius: 50%; flex-shrink: 0;"></div>
                <div class="skeleton-shimmer" style="width: 220px; height: 28px; border-radius: 14px;"></div>
            </div>

            <!-- 2 columns of 3 rounded cards (exact matching user's mockup) -->
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 24px;">
                ${[1,2,3,4,5,6].map(() => `
                    <div class="skeleton-shimmer" style="height: 90px; border-radius: 14px;"></div>
                `).join('')}
            </div>

            <!-- 1 large wide card -->
            <div class="skeleton-shimmer" style="width: 100%; height: 160px; border-radius: 14px; margin-bottom: 20px;"></div>

            <!-- 1 pill bar -->
            <div class="skeleton-shimmer" style="width: 180px; height: 24px; border-radius: 12px; margin-bottom: 16px;"></div>

            <!-- 1 large bottom card -->
            <div class="skeleton-shimmer" style="width: 100%; height: 200px; border-radius: 14px;"></div>
        </div>
    `;
}

export function renderSectionSkeleton(container, path) {
    if (!container) return;
    let html = '';
    if (path === '/profile') {
        html = getProfileSkeletonHTML();
    } else if (path === '/library') {
        html = getLibrarySkeletonHTML();
    } else if (path === '/store') {
        html = getStoreSkeletonHTML();
    } else if (path === '/game') {
        html = getGameDetailSkeletonHTML();
    } else {
        html = getGenericSkeletonHTML();
    }
    container.innerHTML = html;
}

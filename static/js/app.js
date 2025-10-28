class FlightSearchApp {
    constructor() {
        this.currentTab = 'search';
        this.initEventListeners();
        this.setDefaultDates();
        this.showTab('search');
    }

    initEventListeners() {
        const searchForm = document.getElementById('searchForm');
        const trackingForm = document.getElementById('trackingForm');
        
        searchForm.addEventListener('submit', (e) => this.handleSearch(e));
        trackingForm.addEventListener('submit', (e) => this.handleTracking(e));

        // 機場自動完成
        this.setupAutocomplete('origin');
        this.setupAutocomplete('destination');
        this.setupAutocomplete('trackingOrigin');
        this.setupAutocomplete('trackingDestination');

        // 標籤切換
        document.querySelectorAll('.nav-tab').forEach(tab => {
            tab.addEventListener('click', (e) => {
                const tabName = e.target.dataset.tab;
                this.showTab(tabName);
            });
        });
    }

    showTab(tabName) {
        // 更新活躍標籤
        document.querySelectorAll('.nav-tab').forEach(tab => {
            tab.classList.toggle('active', tab.dataset.tab === tabName);
        });

        // 顯示對應內容
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.toggle('active', content.id === `${tabName}Tab`);
        });

        this.currentTab = tabName;
        
        // 隱藏所有結果區域
        this.hideElement('results');
        this.hideElement('trackingResults');
        this.hideElement('error');
        this.hideElement('trackingError');
    }

    setDefaultDates() {
        const today = new Date().toISOString().split('T')[0];
        const nextWeek = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
        
        document.getElementById('departureDate').value = nextWeek;
        document.getElementById('departureDate').min = today;
        document.getElementById('returnDate').min = today;
    }

    setupAutocomplete(fieldId) {
        const input = document.getElementById(fieldId);
        if (!input) return;

        const suggestions = document.getElementById(fieldId + 'Suggestions');
        if (!suggestions) return;

        input.addEventListener('input', (e) => {
            const query = e.target.value.trim();
            if (query.length >= 2) {
                this.searchAirports(query, suggestions);
            } else {
                suggestions.style.display = 'none';
            }
        });

        input.addEventListener('focus', () => {
            if (suggestions.children.length > 0) {
                suggestions.style.display = 'block';
            }
        });

        document.addEventListener('click', (e) => {
            if (!input.contains(e.target) && !suggestions.contains(e.target)) {
                suggestions.style.display = 'none';
            }
        });
    }

    async searchAirports(query, suggestionsContainer) {
        try {
            const response = await fetch(`/api/airports/search?q=${encodeURIComponent(query)}`);
            const data = await response.json();

            if (data.success && data.data.length > 0) {
                this.showSuggestions(data.data, suggestionsContainer);
            } else {
                suggestionsContainer.style.display = 'none';
            }
        } catch (error) {
            console.error('搜尋機場失敗:', error);
        }
    }

    showSuggestions(airports, container) {
        container.innerHTML = '';
        airports.forEach(airport => {
            const div = document.createElement('div');
            div.className = 'suggestion-item';
            div.textContent = `${airport.code} - ${airport.name} (${airport.city})`;
            div.addEventListener('click', () => {
                const input = container.previousElementSibling;
                input.value = airport.code;
                container.style.display = 'none';
            });
            container.appendChild(div);
        });
        container.style.display = 'block';
    }

    // 即時航班搜尋
    async handleSearch(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const params = new URLSearchParams(formData);
        
        console.log('🔍 發送搜尋請求:', params.toString());
        
        this.hideElement('results');
        this.hideElement('error');
        this.showElement('loading');

        try {
            const response = await fetch(`/api/flights/search?${params}`);
            const data = await response.json();

            console.log('📊 收到搜尋響應:', data);

            if (!response.ok) {
                throw new Error(data.error || '搜尋失敗');
            }

            this.displayResults(data);
        } catch (error) {
            console.error('❌ 搜尋錯誤:', error);
            this.showError(error.message);
        } finally {
            this.hideElement('loading');
        }
    }

    // 價格追蹤功能
    async handleTracking(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const origin = formData.get('trackingOrigin');
        const destination = formData.get('trackingDestination');
        const weeks = formData.get('weeks') || 12;
        
        this.hideElement('trackingResults');
        this.hideElement('trackingError');
        this.showElement('trackingLoading');

        try {
            const params = new URLSearchParams({
                origin,
                destination,
                weeks
            });

            console.log('🔍 發送價格追蹤請求:', params.toString());
            
            const response = await fetch(`/api/flights/track-prices?${params}`);
            const data = await response.json();

            console.log('📊 收到價格追蹤響應:', data);

            if (!response.ok) {
                throw new Error(data.error || '價格追蹤失敗');
            }

            this.displayTrackingResults(data);
        } catch (error) {
            console.error('❌ 價格追蹤錯誤:', error);
            this.showTrackingError(error.message);
        } finally {
            this.hideElement('trackingLoading');
        }
    }

    displayResults(data) {
        console.log('🎯 開始顯示結果:', data);
        
        const resultsDiv = document.getElementById('results');
        const countDiv = document.getElementById('resultsCount');
        const flightsDiv = document.getElementById('flightsList');

        // 檢查元素是否存在
        if (!resultsDiv || !countDiv || !flightsDiv) {
            console.error('❌ 找不到必要的DOM元素');
            return;
        }

        // 清空之前的结果
        flightsDiv.innerHTML = '';

        if (!data.success) {
            flightsDiv.innerHTML = `
                <div style="text-align: center; padding: 40px; color: #dc3545;">
                    <i class="fas fa-exclamation-triangle" style="font-size: 3rem; margin-bottom: 20px;"></i>
                    <h3>搜尋失敗</h3>
                    <p>${data.error || '未知錯誤'}</p>
                </div>
            `;
        } else if (!data.data || data.data.length === 0) {
            countDiv.textContent = '找到 0 個航班';
            flightsDiv.innerHTML = `
                <div style="text-align: center; padding: 40px; color: #666;">
                    <i class="fas fa-plane-slash" style="font-size: 3rem; margin-bottom: 20px;"></i>
                    <h3>沒有找到符合條件的航班</h3>
                    <p>請嘗試調整搜尋條件</p>
                </div>
            `;
        } else {
            countDiv.textContent = `找到 ${data.count || data.data.length} 個航班`;
            console.log(`📈 顯示 ${data.data.length} 個航班`);
            
            data.data.forEach((flight, index) => {
                console.log(`✈️ 航班 ${index + 1}:`, flight);
                try {
                    const flightCard = this.createFlightCard(flight);
                    flightsDiv.innerHTML += flightCard;
                } catch (error) {
                    console.error(`❌ 創建航班卡片 ${index + 1} 失敗:`, error);
                }
            });
        }

        this.showElement('results');
        console.log('✅ 結果顯示完成');
    }

    displayTrackingResults(data) {
        const resultsDiv = document.getElementById('trackingResults');
        const analysis = data.data;
        
        console.log('📈 顯示價格分析:', analysis);
        
        resultsDiv.innerHTML = this.createTrackingAnalysis(analysis);
        this.showElement('trackingResults');
    }

    createTrackingAnalysis(analysis) {
        const bestDate = new Date(analysis.best_date).toLocaleDateString('zh-TW');
        
        return `
            <div class="analysis-summary">
                <h3 style="margin-bottom: 20px; text-align: center;"><i class="fas fa-chart-bar"></i> 價格分析摘要</h3>
                <div class="summary-grid">
                    <div class="summary-item">
                        <h3>最低價格</h3>
                        <div class="value">$${this.formatPrice(analysis.min_price)}</div>
                    </div>
                    <div class="summary-item">
                        <h3>平均價格</h3>
                        <div class="value">$${this.formatPrice(analysis.avg_price)}</div>
                    </div>
                    <div class="summary-item">
                        <h3>最高價格</h3>
                        <div class="value">$${this.formatPrice(analysis.max_price)}</div>
                    </div>
                    <div class="summary-item">
                        <h3>最佳出發</h3>
                        <div class="value">${bestDate}</div>
                    </div>
                </div>
                <div class="recommendation" style="background: rgba(255, 255, 255, 0.2); padding: 15px; border-radius: 8px; margin-top: 15px; text-align: center;">
                    <strong>💡 ${analysis.recommendation || '建議根據價格趨勢選擇出發時間'}</strong>
                </div>
            </div>

            <div style="background: #f8f9fa; border-radius: 10px; padding: 20px; margin: 20px 0;">
                <h4><i class="fas fa-history"></i> 價格時間軸</h4>
                <div style="max-height: 400px; overflow-y: auto;">
                    ${analysis.data_points ? analysis.data_points.map(point => this.createTimelineItem(point)).join('') : '沒有價格數據'}
                </div>
            </div>

            <div style="background: #e8f5e8; border: 1px solid #c3e6cb; border-radius: 8px; padding: 15px; color: #155724;">
                <i class="fas fa-check-circle"></i>
                <strong> 分析完成！</strong>
                <p style="margin: 5px 0 0 0;">已分析 ${analysis.track_weeks} 週的價格數據，建議您在 ${bestDate} 附近出發可獲得最優價格。</p>
            </div>
        `;
    }

    createTimelineItem(point) {
        const date = new Date(point.date).toLocaleDateString('zh-TW');
        const isBestPrice = point.price === point.min_price;
        
        return `
            <div style="display: flex; justify-content: space-between; align-items: center; padding: 12px; background: white; border-radius: 8px; margin: 8px 0; border-left: 4px solid ${isBestPrice ? '#28a745' : '#667eea'};">
                <div>
                    <div style="font-weight: 600; color: #333;">${date}</div>
                    <div style="color: #666; font-size: 0.9rem;">第 ${point.week} 週</div>
                </div>
                <div style="font-size: 1.2rem; font-weight: 700; color: #e74c3c;">
                    $${this.formatPrice(point.price)}
                </div>
            </div>
        `;
    }

    createFlightCard(flight) {
        console.log('🎫 創建航班卡片:', flight);
        
        const departureTime = flight.departure ? new Date(flight.departure).toLocaleTimeString('zh-TW', { 
            hour: '2-digit', minute: '2-digit' 
        }) : '未知';
        
        const arrivalTime = flight.arrival ? new Date(flight.arrival).toLocaleTimeString('zh-TW', { 
            hour: '2-digit', minute: '2-digit' 
        }) : '未知';

        const airline = flight.airline || '未知航空公司';
        const stops = flight.stops || 0;
        const price = flight.price || 0;
        const currency = flight.currency || 'TWD';

        return `
            <div class="flight-card">
                <div class="flight-info">
                    <div class="flight-route">
                        <div class="flight-airports">
                            ${flight.from?.code || '未知'} → ${flight.to?.code || '未知'}
                        </div>
                        <div class="flight-duration">
                            ${this.formatDuration(flight.duration)}
                        </div>
                    </div>
                    <div class="flight-details">
                        <span><i class="fas fa-plane"></i> ${airline}</span>
                        <span><i class="fas fa-clock"></i> ${departureTime} - ${arrivalTime}</span>
                        <span><i class="fas fa-stopwatch"></i> ${stops} 次停靠</span>
                        ${flight.flightNumber ? `<span><i class="fas fa-ticket-alt"></i> ${flight.flightNumber}</span>` : ''}
                    </div>
                </div>
                <div class="flight-price">
                    <div class="price">${this.formatPrice(price)}</div>
                    <div class="currency">${currency}</div>
                </div>
            </div>
        `;
    }

    formatDuration(duration) {
        if (!duration) return '未知時長';
        
        const match = duration.match(/PT(?:(\d+)H)?(?:(\d+)M)?/);
        if (!match) return duration;
        
        const hours = match[1] ? parseInt(match[1]) : 0;
        const minutes = match[2] ? parseInt(match[2]) : 0;
        
        let result = '';
        if (hours > 0) result += `${hours}小時`;
        if (minutes > 0) result += `${minutes}分鐘`;
        return result || '0分鐘';
    }

    formatPrice(price) {
        if (!price) return '0';
        return new Intl.NumberFormat('zh-TW').format(Math.round(price));
    }

    showError(message) {
        const errorDiv = document.getElementById('error');
        const errorMessage = document.getElementById('errorMessage');
        
        errorMessage.textContent = message;
        this.showElement('error');
    }

    showTrackingError(message) {
        const errorDiv = document.getElementById('trackingError');
        const errorMessage = document.getElementById('trackingErrorMessage');
        
        errorMessage.textContent = message;
        this.showElement('trackingError');
    }

    showElement(id) {
        const element = document.getElementById(id);
        if (element) {
            element.classList.remove('hidden');
        }
    }

    hideElement(id) {
        const element = document.getElementById(id);
        if (element) {
            element.classList.add('hidden');
        }
    }
}

// 初始化應用
document.addEventListener('DOMContentLoaded', () => {
    new FlightSearchApp();
    console.log('🚀 航班搜尋應用已初始化');
});
class FlightSearchApp {
    constructor() {
        this.currentTab = 'search';
        this.initEventListeners();
        this.setDefaultDates();
        this.showTab('search');
        this.initCurrencyCalculator(); 
        this.initTimeDiffCalculator();
        this.initAttractionsSearch();
    }

    initEventListeners() {
        const searchForm = document.getElementById('searchForm');
        const trackingForm = document.getElementById('trackingForm');
        const timeDiffForm = document.getElementById('timeDiffForm'); 
        
        searchForm.addEventListener('submit', (e) => this.handleSearch(e));
        trackingForm.addEventListener('submit', (e) => this.handleTracking(e));

        // 時差表單提交處理
        if (timeDiffForm) {
            timeDiffForm.addEventListener('submit', (e) => this.handleTimeDiffCalculation(e));
        }

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

   async geocodeLocation(query) {
        try {
            console.log('🗺️ 地理編碼搜尋:', query);
            // 使用免費的 Nominatim API (OpenStreetMap)
            const response = await fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=1`);
            const data = await response.json();
            
            if (data && data.length > 0) {
                console.log('📍 找到位置:', data[0].display_name);
                return {
                    lat: parseFloat(data[0].lat),
                    lng: parseFloat(data[0].lon),
                    displayName: data[0].display_name
                };
            }
            console.warn('❌ 找不到位置:', query);
            return null;
        } catch (error) {
            console.error('地理編碼錯誤:', error);
            return null;
        }
    }

    // 初始化景點搜尋功能
    initAttractionsSearch() {
        const searchBtn = document.getElementById('searchAttractionsBtn');
        
        if (searchBtn) {
            searchBtn.addEventListener('click', (e) => {
                e.preventDefault();
                this.handleAttractionsSearch();
            });
            
            // 也支援 Enter 鍵搜尋
            const latInput = document.getElementById('attractionLat');
            const lngInput = document.getElementById('attractionLng');
            const queryInput = document.getElementById('attractionQuery');
            
            if (latInput && lngInput) {
                [latInput, lngInput, queryInput].forEach(input => {
                    if (input) {
                        input.addEventListener('keypress', (e) => {
                            if (e.key === 'Enter') {
                                e.preventDefault();
                                this.handleAttractionsSearch();
                            }
                        });
                    }
                });
            }
            
            console.log('✅ 景點搜尋功能初始化完成');
        } else {
            console.error('❌ 找不到景點搜尋按鈕');
        }
    }

    // 處理景點搜尋 - 使用地理編碼版本
    async handleAttractionsSearch() {
        console.log('🔍 開始搜尋景點...');
        
        // 獲取輸入值 - 改為地點名稱
        const locationInput = document.getElementById('attractionLocation');
        const radiusSelect = document.getElementById('attractionRadius');
        const queryInput = document.getElementById('attractionQuery');
        const categorySelect = document.getElementById('attractionCategory');
        
        if (!locationInput) {
            console.error('❌ 找不到地點輸入框');
            this.showAttractionsError('系統錯誤：找不到輸入框');
            return;
        }
        
        const locationQuery = locationInput.value.trim();
        
        // 驗證輸入
        if (!locationQuery) {
            this.showAttractionsError('請輸入地點名稱');
            return;
        }

        console.log('📍 搜尋地點:', locationQuery);
        
        // 顯示載入狀態
        this.showAttractionsLoading();
        this.hideAttractionsError();
        this.hideAttractionsResults();

        try {
            // 第一步：地理編碼，將地名轉為經緯度
            const geocodeResult = await this.geocodeLocation(locationQuery);
            
            if (!geocodeResult) {
                throw new Error(`找不到地點 "${locationQuery}"，請嘗試更明確的名稱`);
            }

            console.log('🎯 地理編碼結果:', geocodeResult);
            
            // 第二步：使用經緯度搜尋景點
            const params = new URLSearchParams({
                lat: geocodeResult.lat.toString(),
                lng: geocodeResult.lng.toString(),
                radius: radiusSelect ? radiusSelect.value : '1000'
            });

            if (queryInput && queryInput.value.trim()) {
                params.append('query', queryInput.value.trim());
            }
            if (categorySelect && categorySelect.value && categorySelect.value !== 'all') {
                params.append('category', categorySelect.value);
            }

            const apiUrl = `/api/attractions/search?${params.toString()}`;
            console.log('🌐 發送景點搜尋請求:', apiUrl);
            
            const response = await fetch(apiUrl);
            console.log('📡 API 回應狀態:', response.status);
            
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`HTTP錯誤: ${response.status} - ${errorText}`);
            }

            const data = await response.json();
            console.log('✅ API 回應數據:', data);
            
            this.hideAttractionsLoading();

            if (data.success) {
                // 在結果中顯示地點名稱
                const meta = data.meta || { 
                    radius: radiusSelect ? radiusSelect.value : '1000',
                    location: geocodeResult.displayName
                };
                this.displayAttractionsResults(data.data, meta);
            } else {
                throw new Error(data.message || data.error || '搜尋失敗');
            }
        } catch (error) {
            console.error('❌ 景點搜尋請求失敗:', error);
            this.hideAttractionsLoading();
            this.showAttractionsError(`搜尋失敗: ${error.message}`);
        }
    }

    // 顯示景點搜尋結果
    displayAttractionsResults(attractions, meta) {
        console.log('🎯 顯示景點搜尋結果:', attractions);
        
        const countElement = document.getElementById('attractionsCount');
        const listElement = document.getElementById('attractionsList');
        
        if (!countElement || !listElement) {
            console.error('❌ 找不到景點結果顯示元素');
            return;
        }
        
        // 顯示統計資訊 - 修復：處理空資料情況
        const count = attractions && Array.isArray(attractions) ? attractions.length : 0;
        const radius = meta?.radius || '未知';
        const location = meta?.location || '指定位置'; // 新增：取得地點名稱
        
        // 修改這行：加入地點名稱顯示
        countElement.textContent = `在「${location}」附近找到 ${count} 個景點 (半徑: ${radius} 公尺)`;
        
        // 清空之前的結果
        listElement.innerHTML = '';
        
        if (!attractions || !Array.isArray(attractions) || attractions.length === 0) {
            listElement.innerHTML = `
                <div class="attractions-empty" style="text-align: center; padding: 40px; color: #666;">
                    <i class="fas fa-search-location" style="font-size: 3rem; margin-bottom: 15px;"></i>
                    <h3>在「${location}」附近沒有找到符合條件的景點</h3>  <p>請嘗試：</p>
                    <ul style="text-align: left; margin: 10px 0; display: inline-block;">
                        <li>調整搜尋關鍵字</li>
                        <li>擴大搜尋半徑</li>
                        <li>確認地點名稱是否正確</li>  </ul>
                </div>
            `;
        } else {
            attractions.forEach((attraction, index) => {
                console.log(`🏛️ 景點 ${index + 1}:`, attraction);
                try {
                    const card = this.createAttractionCard(attraction);
                    listElement.appendChild(card);
                } catch (error) {
                    console.error(`❌ 創建景點卡片 ${index + 1} 失敗:`, error);
                    // 創建一個錯誤卡片代替
                    const errorCard = document.createElement('div');
                    errorCard.className = 'attraction-card error';
                    errorCard.innerHTML = `
                        <div style="color: #dc3545; text-align: center; padding: 20px;">
                            <i class="fas fa-exclamation-triangle"></i>
                            <p>無法顯示景點資訊</p>
                        </div>
                    `;
                    listElement.appendChild(errorCard);
                }
            });
        }
        
        this.showAttractionsResults();
    }

    // 創建景點卡片
    createAttractionCard(attraction) {
        const card = document.createElement('div');
        card.className = 'attraction-card';
        card.style.cssText = `
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            border: 1px solid #e0e0e0;
            transition: transform 0.2s, box-shadow 0.2s;
            margin-bottom: 15px;
        `;
        
        card.onmouseover = () => {
            card.style.transform = 'translateY(-2px)';
            card.style.boxShadow = '0 4px 20px rgba(0,0,0,0.15)';
        };
        
        card.onmouseout = () => {
            card.style.transform = 'translateY(0)';
            card.style.boxShadow = '0 2px 10px rgba(0,0,0,0.1)';
        };
        
        // 修復：處理可能為空或未定義的欄位
        const name = attraction.name || '未知名稱';
        const category = attraction.category || attraction.primary_category || '未分類';
        const rating = attraction.rating > 0 ? attraction.rating.toFixed(1) : '無評分';
        const distance = attraction.distance ? Math.round(attraction.distance) : '未知';
        const price = attraction.price > 0 ? '$'.repeat(attraction.price) : '未知';
        
        // 修復：安全處理營業狀態
        let isOpenNow = false;
        let statusText = '營業狀態未知';
        let statusColor = '#6c757d';
        
        if (typeof attraction.is_open_now === 'boolean') {
            isOpenNow = attraction.is_open_now;
            statusText = isOpenNow ? '營業中' : '已休息';
            statusColor = isOpenNow ? '#28a745' : '#dc3545';
        }
        
        // 修復：安全處理其他可能為空的欄位
        const address = attraction.address || attraction.location?.formatted_address || '地址未知';
        const phone = attraction.phone || attraction.contact?.phone || '';
        const website = attraction.website || attraction.contact?.website || '';
        const reviewCount = attraction.review_count || attraction.stats?.review_count || 0;
        
        card.innerHTML = `
            <div class="attraction-header" style="border-bottom: 1px solid #eee; padding-bottom: 15px; margin-bottom: 15px;">
                <h3 class="attraction-name" style="font-size: 1.2rem; font-weight: 600; color: #333; margin: 0 0 8px 0;">${this.escapeHtml(name)}</h3>
                <span class="attraction-category" style="background: #667eea; color: white; padding: 4px 8px; border-radius: 4px; font-size: 0.8rem;">${this.escapeHtml(category)}</span>
            </div>
            <div class="attraction-body">
                <div class="attraction-info" style="display: flex; align-items: center; margin-bottom: 8px;">
                    <i class="fas fa-star" style="color: #ffc107; width: 20px;"></i>
                    <span class="attraction-rating" style="font-weight: 600;">${rating}</span>
                    ${reviewCount > 0 ? `<span class="attraction-reviews" style="color: #666; margin-left: 8px;">(${reviewCount} 則評論)</span>` : ''}
                </div>
                
                <div class="attraction-info" style="display: flex; align-items: center; margin-bottom: 8px;">
                    <i class="fas fa-walking" style="color: #667eea; width: 20px;"></i>
                    <span class="attraction-distance">${distance} 公尺</span>
                </div>
                
                <div class="attraction-info" style="display: flex; align-items: center; margin-bottom: 8px;">
                    <i class="fas fa-dollar-sign" style="color: #28a745; width: 20px;"></i>
                    <span class="attraction-price">${price}</span>
                </div>
                
                <div class="attraction-info" style="display: flex; align-items: center; margin-bottom: 8px;">
                    <i class="fas fa-clock" style="color: ${statusColor}; width: 20px;"></i>
                    <span class="attraction-status" style="color: ${statusColor}; font-weight: 600;">${statusText}</span>
                </div>
                
                ${address && address !== '地址未知' ? `
                <div class="attraction-info" style="display: flex; align-items: flex-start; margin-bottom: 8px;">
                    <i class="fas fa-map-marker-alt" style="color: #e74c3c; width: 20px; margin-top: 2px;"></i>
                    <span class="attraction-address" style="flex: 1;">${this.escapeHtml(address)}</span>
                </div>
                ` : ''}
                
                ${phone ? `
                <div class="attraction-info" style="display: flex; align-items: center; margin-bottom: 8px;">
                    <i class="fas fa-phone" style="color: #007bff; width: 20px;"></i>
                    <span class="attraction-phone">${this.escapeHtml(phone)}</span>
                </div>
                ` : ''}
                
                ${website ? `
                <div class="attraction-info" style="display: flex; align-items: center; margin-bottom: 8px;">
                    <i class="fas fa-globe" style="color: #17a2b8; width: 20px;"></i>
                    <a href="${this.escapeHtml(website)}" target="_blank" class="attraction-website" style="color: #17a2b8; text-decoration: none;">訪問網站</a>
                </div>
                ` : ''}
            </div>
        `;
        
        return card;
    }

    // HTML 轉義工具
    escapeHtml(unsafe) {
        if (!unsafe) return '';
        return unsafe
            .toString()
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }

    // 景點搜尋相關的 UI 控制方法
    showAttractionsLoading() {
        const element = document.getElementById('attractionsLoading');
        if (element) element.classList.remove('hidden');
    }

    hideAttractionsLoading() {
        const element = document.getElementById('attractionsLoading');
        if (element) element.classList.add('hidden');
    }

    showAttractionsError(message) {
        const errorElement = document.getElementById('attractionsError');
        const messageElement = document.getElementById('attractionsErrorMessage');
        
        if (errorElement && messageElement) {
            messageElement.textContent = message;
            errorElement.classList.remove('hidden');
        }
    }

    hideAttractionsError() {
        const element = document.getElementById('attractionsError');
        if (element) element.classList.add('hidden');
    }

    showAttractionsResults() {
        const element = document.getElementById('attractionsResults');
        if (element) element.classList.remove('hidden');
    }

    hideAttractionsResults() {
        const element = document.getElementById('attractionsResults');
        if (element) element.classList.add('hidden');
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
        this.hideAttractionsResults();
        this.hideAttractionsError();
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
        // ★★★ [新增] 搜尋開始前，先隱藏價格建議卡片，避免殘留上次結果 ★★★
        if(document.getElementById('priceAdviceCard')) {
            document.getElementById('priceAdviceCard').classList.add('hidden');
        }
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
        const weatherDiv = document.getElementById('weatherInfo');
        const exchangeDiv = document.getElementById('exchangeInfo');

        // 檢查元素是否存在
        if (!resultsDiv || !countDiv || !flightsDiv || !weatherDiv || !exchangeDiv) {
            console.error('❌ 找不到必要的DOM元素');
            return;
        }

        // 清空之前的結果
        flightsDiv.innerHTML = '';

        if (!data.success) {
            flightsDiv.innerHTML = `
                <div style="text-align: center; padding: 40px; color: #dc3545;">
                    <i class="fas fa-exclamation-triangle" style="font-size: 3rem; margin-bottom: 20px;"></i>
                    <h3>搜尋失敗</h3>
                    <p>${data.error || '未知錯誤'}</p>
                </div>
            `;
            this.hideElement('weatherInfo');
            this.hideElement('exchangeInfo');
        } else if (!data.data || !data.data.flights || data.data.flights.length === 0) {
            countDiv.textContent = '找到 0 個航班';
            flightsDiv.innerHTML = `
                <div style="text-align: center; padding: 40px; color: #666;">
                    <i class="fas fa-plane-slash" style="font-size: 3rem; margin-bottom: 20px;"></i>
                    <h3>沒有找到符合條件的航班</h3>
                    <p>請嘗試調整搜尋條件</p>
                </div>
            `;
            this.hideElement('weatherInfo');
            this.hideElement('exchangeInfo');
        } else {
            const flights = data.data.flights;
            const weatherInfo = data.data.weather;
            const exchangeInfo = data.data.exchange;
            // ★★★ [新增] 取得價格建議資料 ★★★
            const advice = data.data.price_advice;
            
            countDiv.textContent = `找到 ${data.data.meta?.count || flights.length} 個航班`;
            console.log(`📈 顯示 ${flights.length} 個航班`);
            
            // ★★★ [新增] 處理價格建議顯示邏輯 ★★★
            const priceAdviceCard = document.getElementById('priceAdviceCard');
            if (advice && priceAdviceCard) {
                priceAdviceCard.classList.remove('hidden');
                
                // 填入數據
                document.getElementById('adviceText').textContent = advice.advice;
                document.getElementById('adviceCurrent').textContent = '$' + Math.round(advice.current_lowest);
                
                // 處理可能為 0 的歷史數據
                const avgText = advice.history_avg > 0 ? '$' + Math.round(advice.history_avg) : '尚無資料';
                document.getElementById('adviceAvg').textContent = avgText;
                
                const diffText = advice.history_avg > 0 ? advice.diff_percent.toFixed(1) + '%' : '--';
                document.getElementById('adviceDiff').textContent = diffText;
                
                const lowText = advice.history_low > 0 ? '$' + Math.round(advice.history_low) : '--';
                document.getElementById('adviceLow').textContent = lowText;

                // 設定顏色樣式
                let color = '#17a2b8'; // 藍色 (新紀錄/無趨勢)
                let bgColor = '#f0fbfd';
                
                if (advice.trend === 'down') {
                    color = '#28a745'; // 綠色 (降價)
                    bgColor = '#f0fff4';
                } else if (advice.trend === 'up') {
                    color = '#dc3545'; // 紅色 (漲價)
                    bgColor = '#fff0f0';
                } else if (advice.trend === 'stable') {
                    color = '#ffc107'; // 黃色 (持平)
                    bgColor = '#fffbf0';
                }

                priceAdviceCard.style.borderLeftColor = color;
                priceAdviceCard.style.backgroundColor = bgColor;
            }
            // ★★★ [新增結束] ★★★

            // 顯示天氣資訊
            if (weatherInfo) {
                this.displayWeatherInfo(weatherInfo);
                this.showElement('weatherInfo');
            } else {
                this.hideElement('weatherInfo');
            }

            // 顯示匯率資訊
            if (exchangeInfo) {
                this.displayExchangeInfo(exchangeInfo);
                this.showElement('exchangeInfo');
            } else {
                this.hideElement('exchangeInfo');
            }
            
            // 顯示航班列表
            flights.forEach((flight, index) => {
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

    // 顯示天氣資訊 - 修正名稱
    displayWeatherInfo(weatherInfo) {
        console.log('🌤️ 顯示天氣資訊:', weatherInfo);
        
        // ... (保留原本的出發地天氣代碼) ...
        if (weatherInfo.origin_weather) {
            const origin = weatherInfo.origin_weather;
            document.getElementById('originTemp').textContent = `${Math.round(origin.avg_temp)}°C`;
            document.getElementById('originCondition').textContent = origin.condition;
            document.getElementById('originHumidity').textContent = origin.humidity;
            document.getElementById('originWind').textContent = origin.wind_speed;
            document.getElementById('originRain').textContent = origin.chance_of_rain || 0;
            
            const originCityElement = document.querySelector('#originWeather h4');
            if (originCityElement) {
                originCityElement.innerHTML = `<i class="fas fa-plane-departure"></i> ${origin.city} 天氣`;
            }
            
            const originIcon = document.getElementById('originWeatherIcon');
            if (origin.icon && originIcon) {
                originIcon.src = `https:${origin.icon}`;
                originIcon.alt = origin.condition;
            }
        }

        // ... (保留原本的目的地天氣代碼) ...
        if (weatherInfo.destination_weather) {
            const destination = weatherInfo.destination_weather;
            document.getElementById('destinationTemp').textContent = `${Math.round(destination.avg_temp)}°C`;
            document.getElementById('destinationCondition').textContent = destination.condition;
            document.getElementById('destinationHumidity').textContent = destination.humidity;
            document.getElementById('destinationWind').textContent = destination.wind_speed;
            document.getElementById('destinationRain').textContent = destination.chance_of_rain || 0;
            
            const destinationCityElement = document.querySelector('#destinationWeather h4');
            if (destinationCityElement) {
                destinationCityElement.innerHTML = `<i class="fas fa-plane-arrival"></i> ${destination.city} 天氣`;
            }
            
            const destinationIcon = document.getElementById('destinationWeatherIcon');
            if (destination.icon && destinationIcon) {
                destinationIcon.src = `https:${destination.icon}`;
                destinationIcon.alt = destination.condition;
            }
        }

        // ... (保留原本的旅行建議代碼) ...
        if (weatherInfo.travel_advice) {
            document.getElementById('adviceText').textContent = weatherInfo.travel_advice;
        }

        // --- 新增：生成並插入打包清單 ---
        // 1. 移除舊的清單 (如果有)
        const oldList = document.getElementById('dynamicPackingList');
        if (oldList) oldList.remove();

        // 2. 如果有目的地天氣，生成新清單
        if (weatherInfo.destination_weather) {
            const packingItems = this.getPackingList(weatherInfo.destination_weather);
            
            // 建立 HTML 結構
            const packingSection = document.createElement('div');
            packingSection.id = 'dynamicPackingList';
            packingSection.className = 'packing-list-section';
            
            let tagsHtml = packingItems.map(item => `
                <div class="packing-tag">
                    <i class="fas ${item.icon}"></i> ${item.name}
                </div>
            `).join('');

            packingSection.innerHTML = `
                <h4><i class="fas fa-suitcase-rolling"></i> 智慧打包建議 (依據當地天氣)</h4>
                <div class="packing-tags">
                    ${tagsHtml}
                </div>
            `;

            // 3. 插入到天氣區塊的最後面
            const weatherDiv = document.getElementById('weatherInfo');
            if (weatherDiv) {
                weatherDiv.appendChild(packingSection);
            }
        }
    }

    // 顯示匯率資訊
    displayExchangeInfo(exchangeInfo) {
        console.log('💱 顯示匯率資訊:', exchangeInfo);
        
        // 顯示基礎貨幣
        document.getElementById('baseCurrency').textContent = exchangeInfo.base_currency;
        
        // 顯示更新時間
        const lastUpdated = new Date(exchangeInfo.last_updated).toLocaleString('zh-TW');
        document.getElementById('exchangeLastUpdated').textContent = lastUpdated;
        
        // 顯示匯率卡片
        const ratesContainer = document.getElementById('exchangeRates');
        ratesContainer.innerHTML = '';
        
        Object.entries(exchangeInfo.rates).forEach(([currency, rate]) => {
            const currencyNames = {
                'USD': '美元', 'EUR': '歐元', 'JPY': '日圓', 'GBP': '英鎊',
                'CNY': '人民幣', 'KRW': '韓元', 'HKD': '港幣', 'SGD': '新加坡元',
                'TWD': '新台幣'
            };
            
            const rateCard = document.createElement('div');
            rateCard.className = 'exchange-rate-card';
            rateCard.innerHTML = `
                <div class="currency-code">${currency}</div>
                <div class="currency-rate">${rate.toFixed(4)}</div>
                <div class="currency-name">${currencyNames[currency] || currency}</div>
            `;
            ratesContainer.appendChild(rateCard);
        });
    }

    // 初始化貨幣計算機
    initCurrencyCalculator() {
        const calculateBtn = document.getElementById('calculateBtn');
        const swapBtn = document.getElementById('swapCurrencies');
        const amountInput = document.getElementById('calcAmount');

        if (calculateBtn) {
            calculateBtn.addEventListener('click', () => this.handleCurrencyCalculation());
        }

        if (swapBtn) {
            swapBtn.addEventListener('click', () => this.swapCurrencies());
        }

        if (amountInput) {
            amountInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.handleCurrencyCalculation();
                }
            });

            // 實時計算
            amountInput.addEventListener('input', () => {
                if (amountInput.value) {
                    this.handleCurrencyCalculation();
                }
            });
        }

        // 當貨幣選擇改變時自動計算
        const fromSelect = document.getElementById('calcFromCurrency');
        const toSelect = document.getElementById('calcToCurrency');
        
        if (fromSelect) {
            fromSelect.addEventListener('change', () => {
                if (amountInput.value) {
                    this.handleCurrencyCalculation();
                }
            });
        }

        if (toSelect) {
            toSelect.addEventListener('change', () => {
                if (amountInput.value) {
                    this.handleCurrencyCalculation();
                }
            });
        }
    }

    // 交換貨幣
    swapCurrencies() {
        const fromSelect = document.getElementById('calcFromCurrency');
        const toSelect = document.getElementById('calcToCurrency');
        
        const fromValue = fromSelect.value;
        const toValue = toSelect.value;
        
        fromSelect.value = toValue;
        toSelect.value = fromValue;
        
        // 如果金額不為空，重新計算
        const amountInput = document.getElementById('calcAmount');
        if (amountInput.value) {
            this.handleCurrencyCalculation();
        }
    }

    // 處理貨幣計算
    async handleCurrencyCalculation() {
        const amount = parseFloat(document.getElementById('calcAmount').value);
        const fromCurrency = document.getElementById('calcFromCurrency').value;
        const toCurrency = document.getElementById('calcToCurrency').value;
        const resultDiv = document.getElementById('calcResult');
        
        if (!amount || amount <= 0) {
            this.hideElement('calcResult');
            return;
        }
        
        if (fromCurrency === toCurrency) {
            resultDiv.innerHTML = `
                <div style="text-align: center; padding: 20px; color: #666;">
                    <i class="fas fa-info-circle" style="font-size: 2rem; margin-bottom: 10px;"></i>
                    <p>請選擇不同的貨幣進行轉換</p>
                </div>
            `;
            this.showElement('calcResult');
            return;
        }
        
        try {
            const response = await fetch('/api/currency/convert', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    amount: amount,
                    from_currency: fromCurrency,
                    to_currency: toCurrency
                })
            });
            
            const data = await response.json();
            
            if (data.success) {
                const result = data.data;
                this.displayCalculationResult(result);
                this.showElement('calcResult');
            } else {
                this.showCalculationError(data.error || '轉換失敗');
            }
        } catch (error) {
            console.error('貨幣計算錯誤:', error);
            this.showCalculationError('轉換服務暫時不可用');
        }
    }

    // 顯示計算結果
    displayCalculationResult(result) {
        document.getElementById('originalAmountDisplay').textContent = 
            this.formatPrice(result.original_amount);
        document.getElementById('fromCurrencyDisplay').textContent = result.from_currency;
        document.getElementById('convertedAmountDisplay').textContent = 
            this.formatPrice(result.converted_amount);
        document.getElementById('toCurrencyDisplay').textContent = result.to_currency;
        
        // 顯示匯率
        document.getElementById('exchangeRateDisplay').textContent = 
            `1 ${result.from_currency} = ${result.exchange_rate.toFixed(6)} ${result.to_currency}`;
        
        // 顯示反向匯率
        const reverseRate = 1 / result.exchange_rate;
        document.getElementById('reverseRateDisplay').textContent = 
            `1 ${result.to_currency} = ${reverseRate.toFixed(6)} ${result.from_currency}`;
        
        // 顯示更新時間
        const lastUpdated = new Date(result.last_updated).toLocaleString('zh-TW');
        document.getElementById('calcLastUpdated').textContent = lastUpdated;
    }

    // 顯示計算錯誤
    showCalculationError(message) {
        const resultDiv = document.getElementById('calcResult');
        resultDiv.innerHTML = `
            <div style="text-align: center; padding: 20px; color: #dc3545;">
                <i class="fas fa-exclamation-triangle" style="font-size: 2rem; margin-bottom: 10px;"></i>
                <p>${message}</p>
            </div>
        `;
        this.showElement('calcResult');
    }

    // 初始化時差計算機
    initTimeDiffCalculator() {
        console.log('⏰ 時差計算機已初始化');
    }

    // 處理時差計算 - 修復版本
    async handleTimeDiffCalculation(e) {
        e.preventDefault(); 
        
        console.log('⏰ 開始計算時差...');
        
        // 獲取元素
        const timeDiffResultCard = document.getElementById('timeDiffResultCard');
        const timeDiffResultContent = document.getElementById('timeDiffResultContent');
        const timeDiffError = document.getElementById('timeDiffError');
        
        // 輔助函數：隱藏錯誤和結果
        const hideTimeDiffFeedback = () => {
            if (timeDiffError) timeDiffError.classList.add('hidden');
            if (timeDiffResultCard) timeDiffResultCard.classList.add('hidden');
        };
        
        // 輔助函數：顯示錯誤
        const showTimeDiffError = (message) => {
            if (timeDiffError) {
                timeDiffError.classList.remove('hidden');
                document.getElementById('timeDiffErrorMessage').textContent = message;
            }
            if (timeDiffResultCard) timeDiffResultCard.classList.add('hidden');
        };

        hideTimeDiffFeedback();
        
        // 獲取表單數據
        const from = document.getElementById('timeDiffFrom').value.trim();
        const to = document.getElementById('timeDiffTo').value.trim();

        console.log('📍 時區輸入:', { from, to });

        if (!from || !to) {
            showTimeDiffError('請填寫完整的起始和目標時區。');
            return;
        }
        
        try {
            console.log('🌐 發送時差計算請求到 /timediff...');
            
            const formData = new URLSearchParams({
                from: from,
                to: to
            });
            
            console.log('📦 請求資料:', formData.toString());
            
            const response = await fetch('/timediff', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: formData
            });

            console.log('📡 回應狀態:', response.status, response.statusText);
            
            // 檢查回應類型
            const contentType = response.headers.get('content-type');
            if (!contentType || !contentType.includes('application/json')) {
                const text = await response.text();
                console.error('❌ 非 JSON 回應:', text);
                throw new Error('伺服器回應格式錯誤: ' + text);
            }

            const data = await response.json();
            console.log('✅ 回應數據:', data);

            if (!response.ok || data.success === false) {
                const errorMsg = data.error || '計算時差失敗，請檢查時區名稱是否為 Region/City 格式。';
                console.error('❌ 伺服器回報錯誤:', errorMsg);
                showTimeDiffError(errorMsg);
                return;
            }

            // 成功顯示結果
            const { from: resFrom, to: resTo, diffStr, diff } = data;
            
            console.log('🎯 時差計算結果:', { resFrom, resTo, diffStr, diff });
            
            const isFaster = diff > 0;
            const speedText = isFaster ? '快' : '慢';
            const sign = diff >= 0 ? '+' : ''; // 正數或零顯示 + 號

            timeDiffResultContent.innerHTML = `
                <div class="result-display">
                    <div class="location-info">
                        <i class="fas fa-city"></i>
                        <strong>${resFrom}</strong>
                    </div>
                    <i class="fas fa-long-arrow-alt-right result-arrow"></i>
                    <div class="location-info">
                        <i class="fas fa-globe"></i>
                        <strong>${resTo}</strong>
                    </div>
                </div>

                <div class="difference-info">
                    <h3 class="highlight-diff">時差：<span class="diff-value">${sign}${diffStr}</span></h3>
                    <p>（目標時區 <strong>${resTo}</strong> 比起始時區 <strong>${resFrom}</strong> 
                        <span style="font-weight: bold; color: ${isFaster ? '#28a745' : '#dc3545'};">${speedText}</span> 
                        ${Math.abs(diff)} 小時）
                    </p>
                </div>
            `;
            
            if (timeDiffResultCard) {
                timeDiffResultCard.classList.remove('hidden');
                console.log('✅ 時差結果顯示成功');
            }

        } catch (error) {
            console.error('❌ Fetch Error:', error);
            showTimeDiffError('連線錯誤，請檢查網路或後端服務是否正常: ' + error.message);
        }
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

        // --- 新增：判斷紅眼航班 ---
        const isRedEye = this.checkRedEye(flight.departure);
        const redEyeBadge = isRedEye 
            ? `<span class="badge-redeye" title="此航班在深夜起飛"><i class="fas fa-moon"></i> 紅眼航班</span>` 
            : '';

        return `
            <div class="flight-card">
                <div class="flight-info">
                    <div class="flight-route">
                        <div class="flight-airports">
                            ${flight.from?.code || '未知'} → ${flight.to?.code || '未知'}
                        </div>
                        <div class="flight-duration">
                            ${this.formatDuration(flight.duration)} ${redEyeBadge}
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

    // --- 新增功能：檢查是否為紅眼航班 (00:00 - 06:00 起飛) ---
    checkRedEye(departureDateString) {
        if (!departureDateString) return false;
        const date = new Date(departureDateString);
        const hour = date.getHours();
        // 如果是凌晨 0 點到 早上 6 點前，算紅眼
        return hour >= 0 && hour < 6;
    }

    // --- 新增功能：根據天氣生成打包清單 ---
    getPackingList(weather) {
        const items = [
            { icon: 'fa-passport', name: '護照/證件' },
            { icon: 'fa-mobile-alt', name: '充電器/網卡' }
        ];

        if (!weather) return items;

        const temp = weather.avg_temp;
        const condition = weather.condition || '';
        const rainChance = weather.chance_of_rain || 0;

        // 溫度判斷
        if (temp < 10) {
            items.push({ icon: 'fa-snowflake', name: '厚外套/圍巾' });
            items.push({ icon: 'fa-mitten', name: '暖暖包' });
        } else if (temp < 20) {
            items.push({ icon: 'fa-tshirt', name: '薄外套/長袖' });
        } else if (temp > 28) {
            items.push({ icon: 'fa-sun', name: '防曬乳/墨鏡' });
            items.push({ icon: 'fa-fan', name: '手持風扇' });
        }

        // 天氣狀況判斷
        if (condition.includes('Rain') || rainChance > 40) {
            items.push({ icon: 'fa-umbrella', name: '摺疊傘' });
            items.push({ icon: 'fa-shoe-prints', name: '防水鞋' });
        }
        
        return items;
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

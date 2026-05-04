# Журнал работ

<!-- 
// statuses
⌛ - in progress
🧐 - on review
✅ - done
❌ - canceled

// templates
<details>
    <summary>
        List
    </summary>
    <div style="margin-left: 8px">
        <div style="margin-left: 8px">
            List Item
        </div>
    </div> 
</details>

<details style="margin-left: 8px">
    <summary>
        Included List
    </summary>
    <div style="margin-left: 8px">
        <div style="margin-left: 8px">
            List Item
        </div>
    </div> 
</details>

// styles
base unit = 8px 
 -->

<details>
    <summary>
        задокументировать
    </summary>
    <div style="margin-left: 8px">
        <div>
            🧐 особенностей интерфейса (???)
        </div>
        <div>
            🧐 архитектура
        </div>
        <div>
            🧐 договоренности по стилю кода
        </div>
        <div>
            🧐 договоренности по использованию сторонних библ
        </div>
        <div>
            🧐 договоренности по использованию react-query
        </div>
    </div>
</details>

<details>
    <summary>
       <i>app/injections</i>
    </summary>
    <div style="margin-left: 8px">
        <div>
            ✅ axios
        </div>
        <div>
            ✅ react-query
        </div>
    </div>
</details>

<details>
    <summary>
       <i>core/integrations</i>
    </summary>
    <details style="margin-left: 8px">
        <summary>
            <i>axios</i>
        </summary>
        <div style="margin-left: 8px">
            <div>
                ✅ реализовать <i>AxiosClientRegistry</i> с in-memory storage
            </div>
            <div>
                ✅ подключить контекст <i>AxiosContext</i> для registry
            </div>
            <div>
                ✅ реализовать <i>AxiosClientProvider</i> (режимы: <i>registry</i>, <i>clients</i>, <i>client</i>)
            </div>
            <div>
                ✅ реализовать хук <i>useAxiosClient(clientKey?)</i> с fallback на default key
            </div>
            <div>
                ✅ сформировать публичный API <i>core/integrations/axios/index.ts</i>
            </div>
            <div>
                ✅ задокументировать</i>
            </div>
        </div>
        </details>
</details>

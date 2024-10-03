!function(){"use strict";var e,l=window.location,s=window.document,u=s.getElementById("plausible"),p=u.getAttribute("data-api")||(e=(e=u).src.split("/"),d=e[0],e=e[2],d+"//"+e+"/api/event");function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function t(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return c("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return c(null,t);try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var a=u&&u.getAttribute("data-include"),n=u&&u.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(i),n=n&&n.split(",").some(i);if(!a||n)return c("exclusion rule",t)}function i(e){return l.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},n=(a.n=e,a.u=l.href,a.d=u.getAttribute("data-domain"),a.r=s.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),u.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),r=a.p||{},o=(n.forEach(function(e){var t=e.replace("event-",""),e=u.getAttribute(e);r[t]=r[t]||e}),a.p=r,new XMLHttpRequest);o.open("POST",p,!0),o.setRequestHeader("Content-Type","text/plain"),o.send(JSON.stringify(a)),o.onreadystatechange=function(){4===o.readyState&&t&&t.callback&&t.callback({status:o.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=t;for(var n,i=0;i<a.length;i++)t.apply(this,a[i]);function r(){n!==l.pathname&&(n=l.pathname,t("pageview"))}var o,d=window.history;d.pushState&&(o=d.pushState,d.pushState=function(){o.apply(this,arguments),r()},window.addEventListener("popstate",r)),"prerender"===s.visibilityState?s.addEventListener("visibilitychange",function(){n||"visible"!==s.visibilityState||r()}):r();var f=1;function v(e){var t,a,n,i,r;function o(){n||(n=!0,window.location=a.href)}"auxclick"===e.type&&e.button!==f||((t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target))&&t.href&&t.href.split("?")[0],(r=t)&&r.href&&r.host&&r.host!==l.host&&(r=e,e={name:"Outbound Link: Click",props:{url:(a=t).href}},n=!1,!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(r,a)?((i={props:e.props}).revenue=e.revenue,plausible(e.name,i)):((i={props:e.props,callback:o}).revenue=e.revenue,plausible(e.name,i),setTimeout(o,5e3),r.preventDefault())))}s.addEventListener("click",v),s.addEventListener("auxclick",v)}();
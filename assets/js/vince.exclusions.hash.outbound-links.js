!function(){"use strict";var l=window.location,p=window.document,d=p.currentScript,u=d.getAttribute("data-api")||new URL(d.src).origin+"/api/event";function f(e){console.warn("Ignoring Event: "+e)}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return f("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return f("localStorage flag")}catch(e){}var n=d&&d.getAttribute("data-include"),i=d&&d.getAttribute("data-exclude");if("pageview"===e){var r=!n||n&&n.split(",").some(s),a=i&&i.split(",").some(s);if(!r||a)return f("exclusion rule")}var o={};o.n=e,o.u=l.href,o.d=d.getAttribute("data-domain"),o.r=p.referrer||null,o.w=window.innerWidth,t&&t.meta&&(o.m=JSON.stringify(t.meta)),t&&t.props&&(o.p=t.props),o.h=1;var c=new XMLHttpRequest;c.open("POST",u,!0),c.setRequestHeader("Content-Type","text/plain"),c.send(JSON.stringify(o)),c.onreadystatechange=function(){4===c.readyState&&t&&t.callback&&t.callback()}}function s(e){var t=l.pathname;return(t+=l.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n,i=0;i<t.length;i++)e.apply(this,t[i]);function r(){n=l.pathname,e("pageview")}window.addEventListener("hashchange",r),"prerender"===p.visibilityState?p.addEventListener("visibilitychange",function(){n||"visible"!==p.visibilityState||r()}):r();var s=1;function a(e){if("auxclick"!==e.type||e.button===s){var t,n,i,r,a,o=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target);o&&o.href&&o.href.split("?")[0];if((a=o)&&a.href&&a.host&&a.host!==l.host)return t=e,i={name:"Outbound Link: Click",props:{url:(n=o).href}},r=!1,void(!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),i=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&i}}(t,n)?vince(i.name,{props:i.props}):(vince(i.name,{props:i.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){r||(r=!0,window.location=n.href)}}p.addEventListener("click",a),p.addEventListener("auxclick",a)}();
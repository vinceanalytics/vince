!function(){"use strict";var s=window.location,o=window.document,c=o.currentScript,p=c.getAttribute("data-api")||new URL(c.src).origin+"/api/event";function u(t){console.warn("Ignoring Event: "+t)}function t(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(s.hostname)||"file:"===s.protocol)return u("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return u("localStorage flag")}catch(t){}var n={};n.n=t,n.u=s.href,n.d=c.getAttribute("data-domain"),n.r=o.referrer||null,n.w=window.innerWidth,e&&e.meta&&(n.m=JSON.stringify(e.meta)),e&&e.props&&(n.p=e.props);var i=c.getAttributeNames().filter(function(t){return"event-"===t.substring(0,6)}),r=n.p||{};i.forEach(function(t){var e=t.replace("event-",""),n=c.getAttribute(t);r[e]=r[e]||n}),n.p=r;var a=new XMLHttpRequest;a.open("POST",p,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(n)),a.onreadystatechange=function(){4===a.readyState&&e&&e.callback&&e.callback()}}}var e=window.vince&&window.vince.q||[];window.vince=t;for(var n,i=0;i<e.length;i++)t.apply(this,e[i]);function r(){n!==s.pathname&&(n=s.pathname,t("pageview"))}var a,l=window.history;l.pushState&&(a=l.pushState,l.pushState=function(){a.apply(this,arguments),r()},window.addEventListener("popstate",r)),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){n||"visible"!==o.visibilityState||r()}):r();var d=1;function f(t){if("auxclick"!==t.type||t.button===d){var e,n,i,r,a,o=function(t){for(;t&&(void 0===t.tagName||(!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase())||!t.href);)t=t.parentNode;var e;return t}(t.target);o&&o.href&&o.href.split("?")[0];if((a=o)&&a.href&&a.host&&a.host!==s.host)return e=t,i={name:"Outbound Link: Click",props:{url:(n=o).href}},r=!1,void(!function(t,e){if(!t.defaultPrevented){var n=!e.target||e.target.match(/^_(self|parent|top)$/i),i=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type;return n&&i}}(e,n)?vince(i.name,{props:i.props}):(vince(i.name,{props:i.props,callback:c}),setTimeout(c,5e3),e.preventDefault()))}function c(){r||(r=!0,window.location=n.href)}}o.addEventListener("click",f),o.addEventListener("auxclick",f)}();
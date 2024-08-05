!function(){"use strict";var t,l=window.location,s=window.document,p=s.getElementById("plausible"),u=p.getAttribute("data-api")||(t=(t=p).src.split("/"),d=t[0],t=t[2],d+"//"+t+"/api/event");function c(t,e){t&&console.warn("Ignoring Event: "+t),e&&e.callback&&e.callback()}function e(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return c("localhost",e);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return c(null,e);try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",e)}catch(t){}var a=p&&p.getAttribute("data-include"),n=p&&p.getAttribute("data-exclude");if("pageview"===t){a=!a||a.split(",").some(i),n=n&&n.split(",").some(i);if(!a||n)return c("exclusion rule",e)}function i(t){return l.pathname.match(new RegExp("^"+t.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},n=(a.n=t,a.u=l.href,a.d=p.getAttribute("data-domain"),a.r=s.referrer||null,e&&e.meta&&(a.m=JSON.stringify(e.meta)),e&&e.props&&(a.p=e.props),p.getAttributeNames().filter(function(t){return"event-"===t.substring(0,6)})),r=a.p||{},o=(n.forEach(function(t){var e=t.replace("event-",""),t=p.getAttribute(t);r[e]=r[e]||t}),a.p=r,new XMLHttpRequest);o.open("POST",u,!0),o.setRequestHeader("Content-Type","text/plain"),o.send(JSON.stringify(a)),o.onreadystatechange=function(){4===o.readyState&&e&&e.callback&&e.callback({status:o.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=e;for(var n,i=0;i<a.length;i++)e.apply(this,a[i]);function r(){n!==l.pathname&&(n=l.pathname,e("pageview"))}var o,d=window.history;d.pushState&&(o=d.pushState,d.pushState=function(){o.apply(this,arguments),r()},window.addEventListener("popstate",r)),"prerender"===s.visibilityState?s.addEventListener("visibilitychange",function(){n||"visible"!==s.visibilityState||r()}):r();var f=1;function w(t){var e,a,n,i,r;function o(){n||(n=!0,window.location=a.href)}"auxclick"===t.type&&t.button!==f||((e=function(t){for(;t&&(void 0===t.tagName||!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase()||!t.href);)t=t.parentNode;var e;return t}(t.target))&&e.href&&e.href.split("?")[0],(r=e)&&r.href&&r.host&&r.host!==l.host&&(r=t,t={name:"Outbound Link: Click",props:{url:(a=e).href}},n=!1,!function(t,e){if(!t.defaultPrevented)return e=!e.target||e.target.match(/^_(self|parent|top)$/i),t=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type,e&&t}(r,a)?(i={props:t.props},plausible(t.name,i)):(i={props:t.props,callback:o},plausible(t.name,i),setTimeout(o,5e3),r.preventDefault())))}s.addEventListener("click",w),s.addEventListener("auxclick",w)}();
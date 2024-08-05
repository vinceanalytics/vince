!function(){"use strict";var e,t,l=window.location,s=window.document,u=s.getElementById("plausible"),c=u.getAttribute("data-api")||(e=(e=u).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function p(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function n(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return p("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return p(null,t);try{if("true"===window.localStorage.plausible_ignore)return p("localStorage flag",t)}catch(e){}var n=u&&u.getAttribute("data-include"),a=u&&u.getAttribute("data-exclude");if("pageview"===e){n=!n||n.split(",").some(r),a=a&&a.split(",").some(r);if(!n||a)return p("exclusion rule",t)}function r(e){var t=l.pathname;return(t+=l.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var n={},a=(n.n=e,n.u=l.href,n.d=u.getAttribute("data-domain"),n.r=s.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),u.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),i=n.p||{},o=(a.forEach(function(e){var t=e.replace("event-",""),e=u.getAttribute(e);i[t]=i[t]||e}),n.p=i,n.h=1,new XMLHttpRequest);o.open("POST",c,!0),o.setRequestHeader("Content-Type","text/plain"),o.send(JSON.stringify(n)),o.onreadystatechange=function(){4===o.readyState&&t&&t.callback&&t.callback({status:o.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=n;for(var r,i=0;i<a.length;i++)n.apply(this,a[i]);function o(){r=l.pathname,n("pageview")}window.addEventListener("hashchange",o),"prerender"===s.visibilityState?s.addEventListener("visibilitychange",function(){r||"visible"!==s.visibilityState||o()}):o();var d=1;function f(e){var t,n,a,r,i;function o(){a||(a=!0,window.location=n.href)}"auxclick"===e.type&&e.button!==d||((t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target))&&t.href&&t.href.split("?")[0],(i=t)&&i.href&&i.host&&i.host!==l.host&&(i=e,e={name:"Outbound Link: Click",props:{url:(n=t).href}},a=!1,!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(i,n)?((r={props:e.props}).revenue=e.revenue,plausible(e.name,r)):((r={props:e.props,callback:o}).revenue=e.revenue,plausible(e.name,r),setTimeout(o,5e3),i.preventDefault())))}s.addEventListener("click",f),s.addEventListener("auxclick",f)}();
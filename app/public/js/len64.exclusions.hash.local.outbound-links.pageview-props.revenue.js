!function(){"use strict";var u=window.location,l=window.document,c=l.currentScript,s=c.getAttribute("data-api")||new URL(c.src).origin+"/api/event";function p(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return p("localStorage flag",t)}catch(e){}var a=c&&c.getAttribute("data-include"),n=c&&c.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(r),n=n&&n.split(",").some(r);if(!a||n)return p("exclusion rule",t)}function r(e){var t=u.pathname;return(t+=u.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},n=(a.n=e,a.u=u.href,a.d=c.getAttribute("data-domain"),a.r=l.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),c.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),i=a.p||{},o=(n.forEach(function(e){var t=e.replace("event-",""),e=c.getAttribute(e);i[t]=i[t]||e}),a.p=i,a.h=1,new XMLHttpRequest);o.open("POST",s,!0),o.setRequestHeader("Content-Type","text/plain"),o.send(JSON.stringify(a)),o.onreadystatechange=function(){4===o.readyState&&t&&t.callback&&t.callback({status:o.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,n=0;n<t.length;n++)e.apply(this,t[n]);function r(){a=u.pathname,e("pageview")}window.addEventListener("hashchange",r),"prerender"===l.visibilityState?l.addEventListener("visibilitychange",function(){a||"visible"!==l.visibilityState||r()}):r();var f=1;function i(e){var t,a,n,r,i;function o(){n||(n=!0,window.location=a.href)}"auxclick"===e.type&&e.button!==f||((t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target))&&t.href&&t.href.split("?")[0],(i=t)&&i.href&&i.host&&i.host!==u.host&&(i=e,e={name:"Outbound Link: Click",props:{url:(a=t).href}},n=!1,!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(i,a)?((r={props:e.props}).revenue=e.revenue,plausible(e.name,r)):((r={props:e.props,callback:o}).revenue=e.revenue,plausible(e.name,r),setTimeout(o,5e3),i.preventDefault())))}l.addEventListener("click",i),l.addEventListener("auxclick",i)}();
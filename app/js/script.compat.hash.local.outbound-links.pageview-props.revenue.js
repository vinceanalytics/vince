!function(){"use strict";var e,t,l=window.location,o=window.document,s=o.getElementById("plausible"),u=s.getAttribute("data-api")||(e=(e=s).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function a(e,t){try{if("true"===window.localStorage.plausible_ignore)return n=t,(a="localStorage flag")&&console.warn("Ignoring Event: "+a),void(n&&n.callback&&n.callback())}catch(e){}var a={},n=(a.n=e,a.u=l.href,a.d=s.getAttribute("data-domain"),a.r=o.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),s.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),r=a.p||{},i=(n.forEach(function(e){var t=e.replace("event-",""),e=s.getAttribute(e);r[t]=r[t]||e}),a.p=r,a.h=1,new XMLHttpRequest);i.open("POST",u,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(a)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var n=window.plausible&&window.plausible.q||[];window.plausible=a;for(var r,i=0;i<n.length;i++)a.apply(this,n[i]);function c(){r=l.pathname,a("pageview")}window.addEventListener("hashchange",c),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){r||"visible"!==o.visibilityState||c()}):c();var p=1;function d(e){var t,a,n,r,i;function o(){n||(n=!0,window.location=a.href)}"auxclick"===e.type&&e.button!==p||((t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target))&&t.href&&t.href.split("?")[0],(i=t)&&i.href&&i.host&&i.host!==l.host&&(i=e,e={name:"Outbound Link: Click",props:{url:(a=t).href}},n=!1,!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(i,a)?((r={props:e.props}).revenue=e.revenue,plausible(e.name,r)):((r={props:e.props,callback:o}).revenue=e.revenue,plausible(e.name,r),setTimeout(o,5e3),i.preventDefault())))}o.addEventListener("click",d),o.addEventListener("auxclick",d)}();
!function(){"use strict";var l=window.location,o=window.document,s=o.currentScript,c=s.getAttribute("data-api")||new URL(s.src).origin+"/api/event";function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var a=s&&s.getAttribute("data-include"),i=s&&s.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(n),i=i&&i.split(",").some(n);if(!a||i)return u("exclusion rule",t)}function n(e){var t=l.pathname;return(t+=l.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},r=(a.n=e,a.u=l.href,a.d=s.getAttribute("data-domain"),a.r=o.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),a.h=1,new XMLHttpRequest);r.open("POST",c,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,i=0;i<t.length;i++)e.apply(this,t[i]);function n(){a=l.pathname,e("pageview")}window.addEventListener("hashchange",n),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){a||"visible"!==o.visibilityState||n()}):n()}();
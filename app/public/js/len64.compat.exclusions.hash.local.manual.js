!function(){"use strict";var e,t,l=window.location,o=window.document,u=o.getElementById("plausible"),s=u.getAttribute("data-api")||(e=(e=u).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function a(e,t){try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var a=u&&u.getAttribute("data-include"),n=u&&u.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(i),n=n&&n.split(",").some(i);if(!a||n)return c("exclusion rule",t)}function i(e){var t=l.pathname;return(t+=l.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},r=(a.n=e,a.u=t&&t.u?t.u:l.href,a.d=u.getAttribute("data-domain"),a.r=o.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),a.h=1,new XMLHttpRequest);r.open("POST",s,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var n=window.plausible&&window.plausible.q||[];window.plausible=a;for(var i=0;i<n.length;i++)a.apply(this,n[i])}();
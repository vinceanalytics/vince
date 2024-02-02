(function(){"use strict";var n,s,r,o=window.location,t=window.document,e=t.currentScript,d=e.getAttribute("data-api")||l(e);function i(e){console.warn("Ignoring Event: "+e)}function l(e){return new URL(e.src).origin+"/api/event"}function a(n,s){try{if(window.localStorage.vince_ignore==="true")return i("localStorage flag")}catch{}var a,r,c,h,m,p,l=e&&e.getAttribute("data-include"),u=e&&e.getAttribute("data-exclude");if(n==="pageview"&&(h=!l||l&&l.split(",").some(f),m=u&&u.split(",").some(f),!h||m))return i("exclusion rule");function f(e){var t=o.pathname;return t.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^.])\*/g,"$1[^\\s/]*")+"/?$"))}a={},a.n=n,a.u=s&&s.u?s.u:o.href,a.d=e.getAttribute("data-domain"),a.r=t.referrer||null,a.w=window.innerWidth,s&&s.meta&&(a.m=JSON.stringify(s.meta)),s&&s.props&&(a.p=s.props),p=e.getAttributeNames().filter(function(e){return e.substring(0,6)==="event-"}),c=a.p||{},p.forEach(function(t){var n=t.replace("event-",""),s=e.getAttribute(t);c[n]=c[n]||s}),a.p=c,r=new XMLHttpRequest,r.open("POST",d,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){r.readyState===4&&s&&s.callback&&s.callback()}}s=window.vince&&window.vince.q||[],window.vince=a;for(n=0;n<s.length;n++)a.apply(this,s[n]);function u(e){for(;e&&(typeof e.tagName=="undefined"||!h(e)||!e.href);)e=e.parentNode;return e}function h(e){return e&&e.tagName&&e.tagName.toLowerCase()==="a"}function m(e,t){if(e.defaultPrevented)return!1;var n=!t.target||t.target.match(/^_(self|parent|top)$/i),s=!(e.ctrlKey||e.metaKey||e.shiftKey)&&e.type==="click";return n&&s}r=1;function c(e){if(e.type==="auxclick"&&e.button!==r)return;var t=u(e.target),n=t&&t.href&&t.href.split("?")[0];if(p(t))return f(e,t,{name:"Outbound Link: Click",props:{url:t.href}})}function f(e,t,n){var s=!1;function o(){s||(s=!0,window.location=t.href)}m(e,t)?(vince(n.name,{props:n.props,callback:o}),setTimeout(o,5e3),e.preventDefault()):vince(n.name,{props:n.props})}t.addEventListener("click",c),t.addEventListener("auxclick",c);function p(e){return e&&e.href&&e.host&&e.host!==o.host}})()
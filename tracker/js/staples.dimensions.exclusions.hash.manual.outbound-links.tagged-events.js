(function(){"use strict";var s,i,r,c,n=window.location,t=window.document,e=t.currentScript,w=e.getAttribute("data-api")||g(e);function o(e){console.warn("Ignoring Event: "+e)}function g(e){return new URL(e.src).origin+"/api/event"}function d(s,i){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(n.hostname)||n.protocol==="file:")return o("localhost");if(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)return;try{if(window.localStorage.vince_ignore==="true")return o("localStorage flag")}catch{}var a,r,c,u,h,f,l=e&&e.getAttribute("data-include"),d=e&&e.getAttribute("data-exclude");if(s==="pageview"&&(u=!l||l&&l.split(",").some(m),h=d&&d.split(",").some(m),!u||h))return o("exclusion rule");function m(e){var t=n.pathname;return t+=n.hash,t.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^.])\*/g,"$1[^\\s/]*")+"/?$"))}a={},a.n=s,a.u=i&&i.u?i.u:n.href,a.d=e.getAttribute("data-domain"),a.r=t.referrer||null,a.w=window.innerWidth,i&&i.meta&&(a.m=JSON.stringify(i.meta)),i&&i.props&&(a.p=i.props),f=e.getAttributeNames().filter(function(e){return e.substring(0,6)==="event-"}),c=a.p||{},f.forEach(function(t){var n=t.replace("event-",""),s=e.getAttribute(t);c[n]=c[n]||s}),a.p=c,a.h=1,r=new XMLHttpRequest,r.open("POST",w,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){r.readyState===4&&i&&i.callback&&i.callback()}}i=window.vince&&window.vince.q||[],window.vince=d;for(s=0;s<i.length;s++)d.apply(this,i[s]);function _(e){for(;e&&(typeof e.tagName=="undefined"||!h(e)||!e.href);)e=e.parentNode;return e}function h(e){return e&&e.tagName&&e.tagName.toLowerCase()==="a"}function y(e,t){if(e.defaultPrevented)return!1;var n=!t.target||t.target.match(/^_(self|parent|top)$/i),s=!(e.ctrlKey||e.metaKey||e.shiftKey)&&e.type==="click";return n&&s}r=1;function f(e){if(e.type==="auxclick"&&e.button!==r)return;var t=_(e.target),n=t&&t.href&&t.href.split("?")[0];if(u(t,0))return;if(j(t))return p(e,t,{name:"Outbound Link: Click",props:{url:t.href}})}function p(e,t,n){var s=!1;function o(){s||(s=!0,window.location=t.href)}y(e,t)?(vince(n.name,{props:n.props,callback:o}),setTimeout(o,5e3),e.preventDefault()):vince(n.name,{props:n.props})}t.addEventListener("click",f),t.addEventListener("auxclick",f);function j(e){return e&&e.href&&e.host&&e.host!==n.host}function l(e){var n,s,i,r,l,c=a(e)?e:e&&e.parentNode,t={name:null,props:{}},o=c&&c.classList;if(!o)return t;for(n=0;n<o.length;n++){if(l=o.item(n),s=l.match(/vince-event-(.+)=(.+)/),!s)continue;i=s[1],r=s[2].replace(/\+/g," "),i.toLowerCase()==="name"?t.name=r:t.props[i]=r}return t}function v(e){var n,s=e.target,t=l(s);if(!t.name)return;e.preventDefault(),n=!1;function o(){n||(n=!0,s.submit())}setTimeout(o,5e3),vince(t.name,{props:t.props,callback:o})}function b(e){return e&&e.tagName&&e.tagName.toLowerCase()==="form"}c=3;function m(e){if(e.type==="auxclick"&&e.button!==r)return;for(var n,s,o,t=e.target,i=0;i<=c;i++){if(!t)break;if(b(t))return;h(t)&&(s=t),a(t)&&(o=t),t=t.parentNode}o&&(n=l(o),s?(n.props.url=s.href,p(e,s,n)):vince(n.name,{props:n.props}))}function a(e){var t,n=e&&e.classList;if(n)for(t=0;t<n.length;t++)if(n.item(t).match(/vince-event-name=(.+)/))return!0;return!1}function u(e,t){return!(!e||t>c)&&(!!a(e)||u(e.parentNode,t+1))}t.addEventListener("submit",v),t.addEventListener("click",m),t.addEventListener("auxclick",m)})()
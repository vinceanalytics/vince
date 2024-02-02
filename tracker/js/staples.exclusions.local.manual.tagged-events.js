(function(){"use strict";var n,s,i,a,c=window.location,e=window.document,t=e.currentScript,y=t.getAttribute("data-api")||b(t);function h(e){console.warn("Ignoring Event: "+e)}function b(e){return new URL(e.src).origin+"/api/event"}function f(n,s){try{if(window.localStorage.vince_ignore==="true")return h("localStorage flag")}catch{}var o,i,l,d,a=t&&t.getAttribute("data-include"),r=t&&t.getAttribute("data-exclude");if(n==="pageview"&&(l=!a||a&&a.split(",").some(u),d=r&&r.split(",").some(u),!l||d))return h("exclusion rule");function u(e){var t=c.pathname;return t.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^.])\*/g,"$1[^\\s/]*")+"/?$"))}o={},o.n=n,o.u=s&&s.u?s.u:c.href,o.d=t.getAttribute("data-domain"),o.r=e.referrer||null,o.w=window.innerWidth,s&&s.meta&&(o.m=JSON.stringify(s.meta)),s&&s.props&&(o.p=s.props),i=new XMLHttpRequest,i.open("POST",y,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(o)),i.onreadystatechange=function(){i.readyState===4&&s&&s.callback&&s.callback()}}s=window.vince&&window.vince.q||[],window.vince=f;for(n=0;n<s.length;n++)f.apply(this,s[n]);function _(e){for(;e&&(typeof e.tagName=="undefined"||!d(e)||!e.href);)e=e.parentNode;return e}function d(e){return e&&e.tagName&&e.tagName.toLowerCase()==="a"}function j(e,t){if(e.defaultPrevented)return!1;var n=!t.target||t.target.match(/^_(self|parent|top)$/i),s=!(e.ctrlKey||e.metaKey||e.shiftKey)&&e.type==="click";return n&&s}a=1;function m(e){if(e.type==="auxclick"&&e.button!==a)return;var t=_(e.target),n=t&&t.href&&t.href.split("?")[0];if(l(t,0))return}function p(e,t,n){var s=!1;function o(){s||(s=!0,window.location=t.href)}j(e,t)?(vince(n.name,{props:n.props,callback:o}),setTimeout(o,5e3),e.preventDefault()):vince(n.name,{props:n.props})}e.addEventListener("click",m),e.addEventListener("auxclick",m);function r(e){var n,s,a,r,l,c=o(e)?e:e&&e.parentNode,t={name:null,props:{}},i=c&&c.classList;if(!i)return t;for(n=0;n<i.length;n++){if(l=i.item(n),s=l.match(/vince-event-(.+)=(.+)/),!s)continue;a=s[1],r=s[2].replace(/\+/g," "),a.toLowerCase()==="name"?t.name=r:t.props[a]=r}return t}function g(e){var n,s=e.target,t=r(s);if(!t.name)return;e.preventDefault(),n=!1;function o(){n||(n=!0,s.submit())}setTimeout(o,5e3),vince(t.name,{props:t.props,callback:o})}function v(e){return e&&e.tagName&&e.tagName.toLowerCase()==="form"}i=3;function u(e){if(e.type==="auxclick"&&e.button!==a)return;for(var n,s,c,t=e.target,l=0;l<=i;l++){if(!t)break;if(v(t))return;d(t)&&(s=t),o(t)&&(c=t),t=t.parentNode}c&&(n=r(c),s?(n.props.url=s.href,p(e,s,n)):vince(n.name,{props:n.props}))}function o(e){var t,n=e&&e.classList;if(n)for(t=0;t<n.length;t++)if(n.item(t).match(/vince-event-name=(.+)/))return!0;return!1}function l(e,t){return!(!e||t>i)&&(!!o(e)||l(e.parentNode,t+1))}e.addEventListener("submit",g),e.addEventListener("click",u),e.addEventListener("auxclick",u)})()
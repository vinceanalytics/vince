(function(){"use strict";var s,i,a,r,l,u,h,m,y,n=window.location,e=window.document,t=e.currentScript,A=t.getAttribute("data-api")||E(t);function _(e){console.warn("Ignoring Event: "+e)}function E(e){return new URL(e.src).origin+"/api/event"}function o(s,o){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(n.hostname)||n.protocol==="file:")return _("localhost");if(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)return;try{if(window.localStorage.vince_ignore==="true")return _("localStorage flag")}catch{}var a,i={};i.n=s,i.u=n.href,i.d=t.getAttribute("data-domain"),i.r=e.referrer||null,i.w=window.innerWidth,o&&o.meta&&(i.m=JSON.stringify(o.meta)),o&&o.props&&(i.p=o.props),i.h=1,a=new XMLHttpRequest,a.open("POST",A,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(i)),a.onreadystatechange=function(){a.readyState===4&&o&&o.callback&&o.callback()}}r=window.vince&&window.vince.q||[],window.vince=o;for(s=0;s<r.length;s++)o.apply(this,r[s]);function d(){y=n.pathname,o("pageview")}window.addEventListener("hashchange",d);function k(){!y&&e.visibilityState==="visible"&&d()}e.visibilityState==="prerender"?e.addEventListener("visibilitychange",k):d();function w(e){for(;e&&(typeof e.tagName=="undefined"||!p(e)||!e.href);)e=e.parentNode;return e}function p(e){return e&&e.tagName&&e.tagName.toLowerCase()==="a"}function S(e,t){if(e.defaultPrevented)return!1;var n=!t.target||t.target.match(/^_(self|parent|top)$/i),s=!(e.ctrlKey||e.metaKey||e.shiftKey)&&e.type==="click";return n&&s}a=1;function g(e){if(e.type==="auxclick"&&e.button!==a)return;var t=w(e.target),n=t&&t.href&&t.href.split("?")[0];if(b(t,0))return;if(O(n))return v(e,t,{name:"File Download",props:{url:n}})}function v(e,t,n){var s=!1;function o(){s||(s=!0,window.location=t.href)}S(e,t)?(vince(n.name,{props:n.props,callback:o}),setTimeout(o,5e3),e.preventDefault()):vince(n.name,{props:n.props})}e.addEventListener("click",g),e.addEventListener("auxclick",g),h=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],u=t.getAttribute("file-types"),l=t.getAttribute("add-file-types"),m=u&&u.split(",")||l&&l.split(",").concat(h)||h;function O(e){if(!e)return!1;var t=e.split(".").pop();return m.some(function(e){return e===t})}function f(e){var n,s,i,a,l,r=c(e)?e:e&&e.parentNode,t={name:null,props:{}},o=r&&r.classList;if(!o)return t;for(n=0;n<o.length;n++){if(l=o.item(n),s=l.match(/vince-event-(.+)=(.+)/),!s)continue;i=s[1],a=s[2].replace(/\+/g," "),i.toLowerCase()==="name"?t.name=a:t.props[i]=a}return t}function x(e){var n,s=e.target,t=f(s);if(!t.name)return;e.preventDefault(),n=!1;function o(){n||(n=!0,s.submit())}setTimeout(o,5e3),vince(t.name,{props:t.props,callback:o})}function C(e){return e&&e.tagName&&e.tagName.toLowerCase()==="form"}i=3;function j(e){if(e.type==="auxclick"&&e.button!==a)return;for(var n,s,o,t=e.target,r=0;r<=i;r++){if(!t)break;if(C(t))return;p(t)&&(s=t),c(t)&&(o=t),t=t.parentNode}o&&(n=f(o),s?(n.props.url=s.href,v(e,s,n)):vince(n.name,{props:n.props}))}function c(e){var t,n=e&&e.classList;if(n)for(t=0;t<n.length;t++)if(n.item(t).match(/vince-event-name=(.+)/))return!0;return!1}function b(e,t){return!(!e||t>i)&&(!!c(e)||b(e.parentNode,t+1))}e.addEventListener("submit",x),e.addEventListener("click",j),e.addEventListener("auxclick",j)})()
(function(){"use strict";var s,i,r,l,d,u,h,m,f,p,O,n=window.location,t=window.document,e=t.currentScript,F=e.getAttribute("data-api")||M(e);function _(e){console.warn("Ignoring Event: "+e)}function M(e){return new URL(e.src).origin+"/api/event"}function a(s,o){try{if(window.localStorage.vince_ignore==="true")return _("localStorage flag")}catch{}var i,a,r,d,u,m,c=e&&e.getAttribute("data-include"),l=e&&e.getAttribute("data-exclude");if(s==="pageview"&&(d=!c||c&&c.split(",").some(h),u=l&&l.split(",").some(h),!d||u))return _("exclusion rule");function h(e){var t=n.pathname;return t.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^.])\*/g,"$1[^\\s/]*")+"/?$"))}i={},i.n=s,i.u=n.href,i.d=e.getAttribute("data-domain"),i.r=t.referrer||null,i.w=window.innerWidth,o&&o.meta&&(i.m=JSON.stringify(o.meta)),o&&o.props&&(i.p=o.props),m=e.getAttributeNames().filter(function(e){return e.substring(0,6)==="event-"}),r=i.p||{},m.forEach(function(t){var n=t.replace("event-",""),s=e.getAttribute(t);r[n]=r[n]||s}),i.p=r,a=new XMLHttpRequest,a.open("POST",F,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(i)),a.onreadystatechange=function(){a.readyState===4&&o&&o.callback&&o.callback()}}r=window.vince&&window.vince.q||[],window.vince=a;for(i=0;i<r.length;i++)a.apply(this,r[i]);function o(){if(l===n.pathname)return;l=n.pathname,a("pageview")}s=window.history,s.pushState&&(p=s.pushState,s.pushState=function(){p.apply(this,arguments),o()},window.addEventListener("popstate",o));function S(){!l&&t.visibilityState==="visible"&&o()}t.visibilityState==="prerender"?t.addEventListener("visibilitychange",S):o();function A(e){for(;e&&(typeof e.tagName=="undefined"||!g(e)||!e.href);)e=e.parentNode;return e}function g(e){return e&&e.tagName&&e.tagName.toLowerCase()==="a"}function x(e,t){if(e.defaultPrevented)return!1;var n=!t.target||t.target.match(/^_(self|parent|top)$/i),s=!(e.ctrlKey||e.metaKey||e.shiftKey)&&e.type==="click";return n&&s}m=1;function b(e){if(e.type==="auxclick"&&e.button!==m)return;var t=A(e.target),n=t&&t.href&&t.href.split("?")[0];if(w(t,0))return;if(C(n))return j(e,t,{name:"File Download",props:{url:n}})}function j(e,t,n){var s=!1;function o(){s||(s=!0,window.location=t.href)}x(e,t)?(vince(n.name,{props:n.props,callback:o}),setTimeout(o,5e3),e.preventDefault()):vince(n.name,{props:n.props})}t.addEventListener("click",b),t.addEventListener("auxclick",b),h=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],d=e.getAttribute("file-types"),u=e.getAttribute("add-file-types"),O=d&&d.split(",")||u&&u.split(",").concat(h)||h;function C(e){if(!e)return!1;var t=e.split(".").pop();return O.some(function(e){return e===t})}function y(e){var n,s,i,a,l,r=c(e)?e:e&&e.parentNode,t={name:null,props:{}},o=r&&r.classList;if(!o)return t;for(n=0;n<o.length;n++){if(l=o.item(n),s=l.match(/vince-event-(.+)=(.+)/),!s)continue;i=s[1],a=s[2].replace(/\+/g," "),i.toLowerCase()==="name"?t.name=a:t.props[i]=a}return t}function E(e){var n,s=e.target,t=y(s);if(!t.name)return;e.preventDefault(),n=!1;function o(){n||(n=!0,s.submit())}setTimeout(o,5e3),vince(t.name,{props:t.props,callback:o})}function k(e){return e&&e.tagName&&e.tagName.toLowerCase()==="form"}f=3;function v(e){if(e.type==="auxclick"&&e.button!==m)return;for(var n,s,o,t=e.target,i=0;i<=f;i++){if(!t)break;if(k(t))return;g(t)&&(s=t),c(t)&&(o=t),t=t.parentNode}o&&(n=y(o),s?(n.props.url=s.href,j(e,s,n)):vince(n.name,{props:n.props}))}function c(e){var t,n=e&&e.classList;if(n)for(t=0;t<n.length;t++)if(n.item(t).match(/vince-event-name=(.+)/))return!0;return!1}function w(e,t){return!(!e||t>f)&&(!!c(e)||w(e.parentNode,t+1))}t.addEventListener("submit",E),t.addEventListener("click",v),t.addEventListener("auxclick",v)})()
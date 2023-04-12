!function(){"use strict";var f=window.location,l=window.document,v=l.currentScript,d=v.getAttribute("data-api")||new URL(v.src).origin+"/api/event";function m(e){console.warn("Ignoring Event: "+e)}function e(e,t){try{if("true"===window.localStorage.vince_ignore)return m("localStorage flag")}catch(e){}var r=v&&v.getAttribute("data-include"),n=v&&v.getAttribute("data-exclude");if("pageview"===e){var a=!r||r&&r.split(",").some(o),i=n&&n.split(",").some(o);if(!a||i)return m("exclusion rule")}function o(e){return f.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=t&&t.u?t.u:f.href,c.d=v.getAttribute("data-domain"),c.r=l.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props);var p=v.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),u=c.p||{};p.forEach(function(e){var t=e.replace("event-",""),r=v.getAttribute(e);u[t]=u[t]||r}),c.p=u;var s=new XMLHttpRequest;s.open("POST",d,!0),s.setRequestHeader("Content-Type","text/plain"),s.send(JSON.stringify(c)),s.onreadystatechange=function(){4===s.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var p=1;function n(e){if("auxclick"!==e.type||e.button===p){var t,r=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),n=r&&r.href&&r.href.split("?")[0];if(!function e(t,r){if(!t||w<r)return!1;if(b(t))return!0;return e(t.parentNode,r+1)}(r,0))return(t=r)&&t.href&&t.host&&t.host!==f.host?u(e,r,{name:"Outbound Link: Click",props:{url:r.href}}):function(e){if(!e)return!1;var t=e.split(".").pop();return s.some(function(e){return e===t})}(n)?u(e,r,{name:"File Download",props:{url:n}}):void 0}}function u(e,t,r){var n=!1;function a(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var r=!t.target||t.target.match(/^_(self|parent|top)$/i),n=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return r&&n}}(e,t)?vince(r.name,{props:r.props}):(vince(r.name,{props:r.props,callback:a}),setTimeout(a,5e3),e.preventDefault())}l.addEventListener("click",n),l.addEventListener("auxclick",n);var a=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],i=v.getAttribute("file-types"),o=v.getAttribute("add-file-types"),s=i&&i.split(",")||o&&o.split(",").concat(a)||a;function g(e){var t=b(e)?e:e&&e.parentNode,r={name:null,props:{}},n=t&&t.classList;if(!n)return r;for(var a=0;a<n.length;a++){var i,o,c=n.item(a).match(/vince-event-(.+)=(.+)/);c&&(i=c[1],o=c[2].replace(/\+/g," "),"name"===i.toLowerCase()?r.name=o:r.props[i]=o)}return r}var w=3;function h(e){if("auxclick"!==e.type||e.button===p){for(var t,r,n,a,i=e.target,o=0;o<=w&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;c(i)&&(t=i),b(i)&&(r=i),i=i.parentNode}r&&(a=g(r),t?(a.props.url=t.href,u(e,t,a)):vince(a.name,{props:a.props}))}}function b(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/vince-event-name=(.+)/))return 1}l.addEventListener("submit",function(e){var t,r=e.target,n=g(r);function a(){t||(t=!0,r.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),vince(n.name,{props:n.props,callback:a}))}),l.addEventListener("click",h),l.addEventListener("auxclick",h)}();
!function(){"use strict";var s=window.location,u=window.document,l=u.currentScript,f=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function v(e){console.warn("Ignoring Event: "+e)}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(s.hostname)||"file:"===s.protocol)return v("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return v("localStorage flag")}catch(e){}var n=l&&l.getAttribute("data-include"),r=l&&l.getAttribute("data-exclude");if("pageview"===e){var i=!n||n&&n.split(",").some(c),a=r&&r.split(",").some(c);if(!i||a)return v("exclusion rule")}var o={};o.n=e,o.u=s.href,o.d=l.getAttribute("data-domain"),o.r=u.referrer||null,o.w=window.innerWidth,t&&t.meta&&(o.m=JSON.stringify(t.meta)),t&&t.props&&(o.p=t.props);var p=new XMLHttpRequest;p.open("POST",f,!0),p.setRequestHeader("Content-Type","text/plain"),p.send(JSON.stringify(o)),p.onreadystatechange=function(){4===p.readyState&&t&&t.callback&&t.callback()}}function c(e){return s.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n,r=0;r<t.length;r++)e.apply(this,t[r]);function i(){n!==s.pathname&&(n=s.pathname,e("pageview"))}var a,o=window.history;function p(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}o.pushState&&(a=o.pushState,o.pushState=function(){a.apply(this,arguments),i()},window.addEventListener("popstate",i)),"prerender"===u.visibilityState?u.addEventListener("visibilitychange",function(){n||"visible"!==u.visibilityState||i()}):i();var c=1;function d(e){if("auxclick"!==e.type||e.button===c){var t=function(e){for(;e&&(void 0===e.tagName||!p(e)||!e.href);)e=e.parentNode;return e}(e.target),n=t&&t.href&&t.href.split("?")[0];if(!function e(t,n){if(!t||x<n)return!1;if(k(t))return!0;return e(t.parentNode,n+1)}(t,0))return function(e){if(!e)return!1;var t=e.split(".").pop();return y.some(function(e){return e===t})}(n)?m(e,t,{name:"File Download",props:{url:n}}):void 0}}function m(e,t,n){var r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:i}),setTimeout(i,5e3),e.preventDefault())}u.addEventListener("click",d),u.addEventListener("auxclick",d);var w=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],g=l.getAttribute("file-types"),h=l.getAttribute("add-file-types"),y=g&&g.split(",")||h&&h.split(",").concat(w)||w;function b(e){var t=k(e)?e:e&&e.parentNode,n={name:null,props:{}},r=t&&t.classList;if(!r)return n;for(var i=0;i<r.length;i++){var a,o,p=r.item(i).match(/vince-event-(.+)=(.+)/);p&&(a=p[1],o=p[2].replace(/\+/g," "),"name"===a.toLowerCase()?n.name=o:n.props[a]=o)}return n}var x=3;function L(e){if("auxclick"!==e.type||e.button===c){for(var t,n,r,i,a=e.target,o=0;o<=x&&a;o++){if((r=a)&&r.tagName&&"form"===r.tagName.toLowerCase())return;p(a)&&(t=a),k(a)&&(n=a),a=a.parentNode}n&&(i=b(n),t?(i.props.url=t.href,m(e,t,i)):vince(i.name,{props:i.props}))}}function k(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}u.addEventListener("submit",function(e){var t,n=e.target,r=b(n);function i(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(i,5e3),vince(r.name,{props:r.props,callback:i}))}),u.addEventListener("click",L),u.addEventListener("auxclick",L)}();
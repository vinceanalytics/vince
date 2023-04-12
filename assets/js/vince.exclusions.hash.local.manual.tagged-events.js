!function(){"use strict";var p=window.location,s=window.document,l=s.currentScript,f=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function v(e){console.warn("Ignoring Event: "+e)}function e(e,t){try{if("true"===window.localStorage.vince_ignore)return v("localStorage flag")}catch(e){}var n=l&&l.getAttribute("data-include"),r=l&&l.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(o),i=r&&r.split(",").some(o);if(!a||i)return v("exclusion rule")}function o(e){var t=p.pathname;return(t+=p.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=t&&t.u?t.u:p.href,c.d=l.getAttribute("data-domain"),c.r=s.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props),c.h=1;var u=new XMLHttpRequest;u.open("POST",f,!0),u.setRequestHeader("Content-Type","text/plain"),u.send(JSON.stringify(c)),u.onreadystatechange=function(){4===u.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n=0;n<t.length;n++)e.apply(this,t[n]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var u=1;function r(e){var t;"auxclick"===e.type&&e.button!==u||((t=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],function e(t,n){if(!t||g<n)return!1;if(w(t))return!0;return e(t.parentNode,n+1)}(t,0))}function d(e,t,n){var r=!1;function a(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:a}),setTimeout(a,5e3),e.preventDefault())}function m(e){var t=w(e)?e:e&&e.parentNode,n={name:null,props:{}},r=t&&t.classList;if(!r)return n;for(var a=0;a<r.length;a++){var i,o,c=r.item(a).match(/vince-event-(.+)=(.+)/);c&&(i=c[1],o=c[2].replace(/\+/g," "),"name"===i.toLowerCase()?n.name=o:n.props[i]=o)}return n}s.addEventListener("click",r),s.addEventListener("auxclick",r);var g=3;function a(e){if("auxclick"!==e.type||e.button===u){for(var t,n,r,a,i=e.target,o=0;o<=g&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;c(i)&&(t=i),w(i)&&(n=i),i=i.parentNode}n&&(a=m(n),t?(a.props.url=t.href,d(e,t,a)):vince(a.name,{props:a.props}))}}function w(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}s.addEventListener("submit",function(e){var t,n=e.target,r=m(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),vince(r.name,{props:r.props,callback:a}))}),s.addEventListener("click",a),s.addEventListener("auxclick",a)}();
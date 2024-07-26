!function(){"use strict";var o=window.location,p=window.document,l=p.currentScript,u=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return a=t,(r="localStorage flag")&&console.warn("Ignoring Event: "+r),void(a&&a.callback&&a.callback())}catch(e){}var r={},a=(r.n=e,r.u=t&&t.u?t.u:o.href,r.d=l.getAttribute("data-domain"),r.r=p.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),l.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),n=r.p||{},i=(a.forEach(function(e){var t=e.replace("event-",""),e=l.getAttribute(e);n[t]=n[t]||e}),r.p=n,new XMLHttpRequest);i.open("POST",u,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(r)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function s(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var c=1;function a(e){var t,r,a,n;if("auxclick"!==e.type||e.button===c)return t=function(e){for(;e&&(void 0===e.tagName||!s(e)||!e.href);)e=e.parentNode;return e}(e.target),r=t&&t.href&&t.href.split("?")[0],!function e(t,r){if(!t||g<r)return!1;if(w(t))return!0;return e(t.parentNode,r+1)}(t,0)&&(a=r)&&(n=a.split(".").pop(),m.some(function(e){return e===n}))?f(e,t,{name:"File Download",props:{url:r}}):void 0}function f(e,t,r){var a,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(a={props:r.props},plausible(r.name,a)):(a={props:r.props,callback:i},plausible(r.name,a),setTimeout(i,5e3),e.preventDefault())}p.addEventListener("click",a),p.addEventListener("auxclick",a);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=l.getAttribute("file-types"),d=l.getAttribute("add-file-types"),m=i&&i.split(",")||d&&d.split(",").concat(n)||n;function v(e){var e=w(e)?e:e&&e.parentNode,t={name:null,props:{}},r=e&&e.classList;if(r)for(var a=0;a<r.length;a++){var n,i=r.item(a).match(/plausible-event-(.+)(=|--)(.+)/);i&&(n=i[1],i=i[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i)}return t}var g=3;function b(e){if("auxclick"!==e.type||e.button===c){for(var t,r,a,n,i=e.target,o=0;o<=g&&i;o++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;s(i)&&(t=i),w(i)&&(r=i),i=i.parentNode}r&&(n=v(r),t?(n.props.url=t.href,f(e,t,n)):((e={}).props=n.props,plausible(n.name,e)))}}function w(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}p.addEventListener("submit",function(e){var t,r=e.target,a=v(r);function n(){t||(t=!0,r.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),e={props:a.props,callback:n},plausible(a.name,e))}),p.addEventListener("click",b),p.addEventListener("auxclick",b)}();